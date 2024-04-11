#![no_std]
#![no_main]
#![allow(async_fn_in_trait)]

use core::str::{self, FromStr};

mod secrets;
use embassy_time::{Duration, Timer};
use rand::RngCore;
use reqwless::{
    client::{HttpClient, TlsConfig, TlsVerify},
    headers::ContentType,
    request::{Method, RequestBuilder},
};
use secrets::{N2Y0_API_KEY, WIFI_PASSPHRASE, WIFI_SSID};

use cyw43_pio::PioSpi;
use defmt::{error, info, panic, warn};
use defmt_rtt as _;
use embassy_executor::Spawner;
use embassy_net::{
    dns::DnsSocket,
    tcp::{
        client::{TcpClient, TcpClientState},
        TcpSocket,
    },
    Config, Stack, StackResources,
};
use embassy_rp::{
    bind_interrupts,
    clocks::RoscRng,
    gpio::{Level, Output},
    peripherals::{DMA_CH0, PIO0},
    pio::{InterruptHandler, Pio},
};
use panic_probe as _;
use static_cell::StaticCell;

bind_interrupts!(struct Irqs {
    PIO0_IRQ_0 => InterruptHandler<PIO0>;
});

type RUNNER = cyw43::Runner<'static, Output<'static>, PioSpi<'static, PIO0, 0, DMA_CH0>>;

#[embassy_executor::task]
async fn wifi_task(runner: RUNNER) -> ! {
    runner.run().await
}

#[embassy_executor::task]
async fn net_task(stack: &'static Stack<cyw43::NetDriver<'static>>) -> ! {
    stack.run().await
}

async fn get_request<'a>(tcp_sock: &mut TcpSocket<'a>) {
    let request = b"GET /test HTTP/1.1\r\nHost: localhost\r\n\r\n";

    match tcp_sock.write(request).await {
        Ok(n) => info!("Wrote {} bytes to socket!", n),
        Err(e) => defmt::panic!("Write error: {:?}", e),
    }

    tcp_sock.flush().await.unwrap();
}

#[embassy_executor::main]
async fn main(spawner: Spawner) {
    info!("Hello World!");

    // init
    let p = embassy_rp::init(Default::default());
    let mut rng = RoscRng;

    // locate cyw34 firmware
    let fw = unsafe { core::slice::from_raw_parts(0x10100000 as *const u8, 230321) };
    let clm = unsafe { core::slice::from_raw_parts(0x10140000 as *const u8, 4752) };

    let pwr = Output::new(p.PIN_23, Level::Low);
    let cs = Output::new(p.PIN_25, Level::High);
    // let mut led = Output::new(p.PIN_0, Level::Low);
    let mut pio = Pio::new(p.PIO0, Irqs);

    let spi = PioSpi::new(
        &mut pio.common,
        pio.sm0,
        pio.irq0,
        cs,
        p.PIN_24,
        p.PIN_29,
        p.DMA_CH0,
    );

    static STATE: StaticCell<cyw43::State> = StaticCell::new();
    let state = STATE.init(cyw43::State::new());

    let (net_device, mut control, runner) = cyw43::new(state, pwr, spi, fw).await;

    spawner.spawn(wifi_task(runner)).unwrap();

    control.init(clm).await;
    control
        .set_power_management(cyw43::PowerManagementMode::PowerSave)
        .await;

    // FW setup complete

    // OUR CODE

    // Initialize Network Stack

    // needed for correct (random) tcp port assignment
    let seed = rng.next_u64();
    let config = Config::dhcpv4(Default::default());

    static STACK: StaticCell<Stack<cyw43::NetDriver<'static>>> = StaticCell::new();
    static RESOURCES: StaticCell<StackResources<5>> = StaticCell::new();

    let stack = STACK.init(Stack::new(
        net_device,
        config,
        RESOURCES.init(StackResources::<5>::new()),
        seed,
    ));

    spawner.spawn(net_task(stack)).unwrap();

    // Connect to Wifi AP
    loop {
        match control.join_wpa2(WIFI_SSID, WIFI_PASSPHRASE).await {
            Ok(_) => break,
            Err(err) => info!("Unable to connect w/ status={}", err.status),
        }
    }

    // Wait for IP assignment
    info!("Connected! ");
    while !stack.is_config_up() {
        Timer::after_millis(100).await;
    }

    if let Some(x) = stack.config_v4() {
        info!("IP addr: {}", x.address);
    } else {
        // should never reach
        // is after stack.is_config_up()
        panic!("Unable to get IPv4 addr!")
    }

    info!("DHCP finished!");
    control.gpio_set(0, true).await;

    // create tcp client

    let mut tls_rx_buffer = [0u8; 4096];
    let mut tls_tx_buffer = [0u8; 4096];

    let tcp_client_state = TcpClientState::<1, 4096, 4096>::new();
    let mut tcp_sock = TcpClient::new(stack, &tcp_client_state);
    let mut dns_sock = DnsSocket::new(stack);
    let tls = TlsConfig::new(
        rng.next_u64(),
        &mut tls_rx_buffer,
        &mut tls_tx_buffer,
        TlsVerify::None,
    );
    let mut client = HttpClient::new_with_tls(&mut tcp_sock, &mut dns_sock, tls);

    // https server
    let url = "https://192.168.1.120/";

    let mut rx_buff = [0u8; 4096];

    let response = client
        .request(Method::GET, &url)
        .await
        .unwrap()
        .send(&mut rx_buff)
        .await
        .unwrap();

    info!("{:?}", rx_buff);
}
