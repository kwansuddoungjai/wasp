use crate::*;
use std::{sync::mpsc, thread::spawn};

#[derive(Clone)]
pub struct Client {
    url: String,
}

impl Client {
    pub fn new(url: &str) -> errors::Result<Self> {
        return Ok(Client {
            url: url.to_owned(),
        });
    }
    pub fn subscribe(&self, ch: mpsc::Sender<String>) {
        // FIXME should not reconnect every time
        let (mut socket, _) = tungstenite::connect(&self.url).unwrap();
        spawn(move || loop {
            match socket.read_message() {
                Ok(msg) => {
                    if msg.to_string() != "" {
                        ch.send(msg.to_string()).unwrap();
                    }
                }
                Err(tungstenite::Error::ConnectionClosed) => {
                    return Ok(());
                }
                Err(e) => {
                    return Err(format!("subscribe err: {}", e));
                }
            }
        });
    }
}

impl PartialEq for Client {
    fn eq(&self, _other: &Self) -> bool {
        todo!()
    }
}

#[cfg(test)]
mod tests {
    use crate::websocket::Client;
    use std::{net::TcpListener, sync::mpsc, thread::spawn};
    use tungstenite::accept;

    #[derive(Clone)]
    struct MockServerMsg {
        msg: String,
        count: u32,
    }

    #[test]
    fn client_new() {
        let url = "ws://localhost:3012";
        mock_server(url, None);
        let client = Client::new(&url).unwrap();
        assert!(client.url == url);
    }

    #[test]
    fn client_subscribe() {
        let url = "ws://localhost:3012";
        let test_msg = "this is test msg";
        mock_server(
            url,
            Some(MockServerMsg {
                msg: test_msg.to_string(),
                count: 3,
            }),
        );
        let client = Client::new(&url).unwrap();
        let (tx, rx): (mpsc::Sender<String>, mpsc::Receiver<String>) = mpsc::channel();
        client.subscribe(tx);
        let mut cnt = 0;
        for msg in rx.iter() {
            assert!(msg == format!("{}: cnt: {}", test_msg, cnt));
            cnt += 1;
        }
    }

    fn mock_server(input_url: &str, response_msg: Option<MockServerMsg>) {
        let ws_prefix = "ws://";
        let url = match input_url.to_string().strip_prefix(ws_prefix) {
            Some(v) => v.to_string(),
            None => input_url.to_string(),
        };
        spawn(move || {
            let server = TcpListener::bind(url).unwrap();
            for stream in server.incoming() {
                let mut socket = accept(stream.unwrap()).unwrap();
                let msg_opt = response_msg.to_owned();
                if !msg_opt.is_none() {
                    let msg = msg_opt.unwrap();
                    for cnt in 0..msg.count {
                        let msg = format!("{}: cnt: {}", msg.msg, cnt);
                        socket
                            .write_message(tungstenite::Message::Text(msg).into())
                            .unwrap();
                    }
                    socket.close(None).unwrap();
                }
            }
        });
    }
}