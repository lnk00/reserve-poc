use actix_web::{App, HttpResponse, HttpServer, Responder, get};
use serde::Serialize;

#[derive(Serialize)]
struct MyResp {
    success: String,
}

#[get("/")]
async fn hello() -> impl Responder {
    HttpResponse::Ok().json(MyResp {
        success: "true".to_string(),
    })
}

#[actix_web::main]
async fn main() -> Result<(), std::io::Error> {
    HttpServer::new(|| App::new().service(hello))
        .bind(("127.0.0.1", 8448))?
        .run()
        .await
}
