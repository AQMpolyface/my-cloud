use anyhow::Result;
use reqwest;
use reqwest::multipart::{Form, Part};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use std::env;
use std::error::Error;
use std::fs::File;
use std::io::copy;
use std::path::Path;
use tokio;
use tokio::io::AsyncReadExt;

#[derive(Debug, Deserialize, Serialize)]
struct FileInfo {
    name: String,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let args: Vec<String> = env::args().collect();
    match args.len() {
        1 => fetch().await?,
        _ => match args[1].as_str() {
            "list" | "l" => fetch().await?,
            "download" | "d" => {
                if args.len() <= 3 {
                    download(&args[2], "/home/polyface/Desktop/transfer/").await?
                } else {
                    download(&args[2], &args[3]).await?
                }
            }
            "upload" | "u" => upload(&args[2]).await?,
            "delete" | "de" => delete(&args[2]).await?,
            _ => fetch().await?,
        },
    }
    Ok(())
}

async fn download(filename: &str, filepath: &str) -> Result<()> {
    let file_path = Path::new(filepath);
    let mut finalpath: String = String::new();
    if file_path.exists() {
        if file_path.is_dir() {
            if filepath.ends_with("/") {
                finalpath = format!("{}{}", filepath, filename);
            } else {
                finalpath = format!("{}/{}", filepath, filename);
            }
        } else if file_path.is_file() {
            finalpath = format!("{}", filepath);
        }
    } else {
        finalpath = format!("{}", filepath)
    }

    println!("{}", finalpath);
    let download_url = format!("{}{}", "https://example.com/api/download/", filename);
    let client = Client::new();
    let response = client
        .get(download_url)
        .header("Cookie", "uuid=my-cookie")
        .send()
        .await?;

    if !response.status().is_success() {
        anyhow::bail!("Failed to download: HTTP {}", response.status());
    }

    let mut file = File::create(finalpath)?;

    // Copy the response body to the file
    copy(&mut response.bytes().await?.as_ref(), &mut file)?;

    Ok(())
}

async fn upload(filename: &str) -> Result<(), Box<dyn Error>> {
    let client = Client::new();

    let mut file = tokio::fs::File::open(filename).await?;
    let mut buffer = Vec::new();
    file.read_to_end(&mut buffer).await?;

    let file_name = Path::new(filename)
        .file_name()
        .unwrap_or_default()
        .to_str()
        .unwrap_or("unknown");

    // Create multipart form using the buffer
    let part = Part::bytes(buffer).file_name(file_name.to_string());
    let form = Form::new().part("file", part);

    // Send the request
    let response = client
        .post("https://example.com/api/upload")
        .header("Cookie", "uuid=my-cookie")
        .multipart(form)
        .send()
        .await?;

    // Check response
    if response.status().is_success() {
        println!("File uploaded successfully!");
        println!("Server response: {}", response.text().await?);
    } else {
        println!("Upload failed with status: {}", response.status());
        println!("Response body: {}", response.text().await?);
    }
    Ok(())
}

async fn delete(filename: &str) -> Result<(), Box<dyn Error>> {
    let delete_url = format!("{}{}", "https://example.com/api/delete/", filename);
    let client = Client::new();
    let response = client
        .get(delete_url)
        .header("Cookie", "uuid=my-cookie")
        .send()
        .await?;
    let body = response.text().await?;

    println!("{}", body);
    Ok(())
}
async fn fetch() -> Result<(), Box<dyn Error>> {
    let list_url = "https://example.com/api/files";
    let client = Client::new();
    let response = client
        .get(list_url)
        .header("Cookie", "uuid=my-cookie")
        .send()
        .await?;
    let body = response.text().await?;
    let files: Vec<FileInfo> = serde_json::from_str(&body)?;
    for file in files {
        println!("{}", file.name);
    }
    Ok(())
}
