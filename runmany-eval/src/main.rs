use std::{error::Error, fs::File, io::{ErrorKind, Read, Write}, path::PathBuf};
use reqwest::{blocking::{multipart, Client, Response}};
use serde_json::Value;

mod evalresult;
use crate::evalresult::{EvalResult, Stringify, EvalReport};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let argv: Vec<String> = std::env::args().collect(); // [0:cmd, 1:expect, 2:url, 3:path, 4:output]
    let argc = std::env::args().count();
    if argc < 4 {
        print_usage();
        return Ok(());
    }
    
    let expect: u8 = match &argv[1].as_str() {
        &"1" | &"real" | &"genuine" => 1,
        &"2" | &"fake" | &"generated" => 2,
        _ => {
            print_usage();
            return Ok(());
        }
    };
    let path = PathBuf::from(&argv[3]);
    let url: &str = &argv[2];
    let report = run_multiple(path, expect, url);
    
    if argc == 4 {
        return Ok(());
    }
    
    let write_path = &argv[4];
    let report_json = match serde_json::to_string(&report) {
        Ok(j) => j,
        Err(_) => String::from("{}")
    };
    write_report(report_json, PathBuf::from(write_path));
    Ok(())
}

fn print_usage() {
    println!("Usage: runmany-eval [expect] [url] [path] [output]\n");
    println!("expect: analysis result to expect. values:\n\t(1,genuine,real)\tgenuine image\n\t(2,generated,fake)\tgenerated image\n");
    println!("url: image upload endpoint, ex. http://localhost:8080/upload\n");
    println!("path: path containing images for analysis\n");
    println!("output: path to write results to. optional");
}

fn write_report(report: String, write_path: PathBuf) {
    let mut outfile = match File::create(write_path) {
        Ok(f) => f,
        Err(_) => {
            println!("Error opening file. Report:\n{}", report);
            return;
        }
    };
    match outfile.write_all(report.as_bytes()) {
        Ok(_) => {},
        Err(_) => {
            println!("Error writing to file. Report:\n{}", report);
            return;
        }
    }
}

fn run_multiple(path: PathBuf, expected_result: u8, url: &str) -> EvalReport {
    let mut results: Vec<EvalResult> = Vec::new();
    let client = Client::new();
    
    let file_paths = match std::fs::read_dir(&path) {
        Ok(paths) => paths,
        Err(_) => return EvalReport::from(Vec::new())
    };
    
    let files_count = std::fs::read_dir(path).unwrap().count();
    println!("Analyzing {} files", files_count);
    for (idx, fpath) in file_paths.enumerate() {
        let file_name = fpath.as_ref().unwrap().file_name().into_string().unwrap(); // of course
        println!("({}/{}) Performing analysis on file {}", (idx + 1), files_count, file_name);
        
        let file = match File::open(fpath.unwrap().path()) {
            Ok(f) => f,
            Err(e) => {
                println!("{:?}\n", e.source());
                results.push(
                    EvalResult::new(expected_result, 0, file_name)
                );
                continue;
            }
        };
        
        let eval = match upload_file(file_name.clone(), file, &client, url) {
            Ok(val) => val,
            Err(e) => {
                println!("{:?}\n", e.source());
                results.push(
                    EvalResult::new(expected_result, 0, file_name)
                );
                continue;
            }
        };
        
        println!("Analysis of file {} returned {}, expected {}\n", file_name, eval, expected_result);
        results.push(
            EvalResult::new(expected_result, eval, file_name)
        );
    }
    
    let report = EvalReport::from(results);
    print_report(&report);
    report
}

fn print_report(report: &EvalReport) {
    println!("expect\tactual\tfile");
    for res in &report.results {
        println!("{}", res.to_string());
    }
    println!("files analyzed:\t{}", report.files_analyzed);
    println!("expected:\t{}", report.expected_result);
    println!("hits:\t\t{}", report.hits);
    println!("misses:\t\t{}", report.misses);
    println!("fails:\t\t{}", report.fails);
    println!("accuracy:\t{}", report.accuracy);
}

fn upload_file(file_name: String, mut file: File, client: &Client, url: &str) -> Result<u8, std::io::Error>{
    let mut buffer = Vec::new();
    file.read_to_end(&mut buffer)?;
    
    let file_ext = file_name.split(".").last().unwrap();
    let mime = match file_ext {
        "png" => "image/png",
        "jpg" | "jpeg" => "image/jpeg",
        _ => return Err(std::io::Error::new(ErrorKind::InvalidData, "Invalid file type"))
    };
    
    let part = multipart::Part::bytes(buffer)
        .file_name(file_name)
        .mime_str(mime).unwrap();
    
    let form = multipart::Form::new().part("image", part);
    
    let server_response = match client.post(url).multipart(form).send() {
        Ok(resp) => resp,
        Err(e) => return Err(std::io::Error::new(ErrorKind::Other, e.to_string()))
    };
    Ok(get_verdict(server_response))
}

fn get_verdict(response: Response) -> u8 {
    let result_plain = response.text().unwrap();
    match result_plain.find("Analysis Failed") {
        Some(_) => return 0,
        None => {},
    };
    
    let json: Value = serde_json::from_str(result_plain.as_str()).unwrap();
    let verdict = json["analysis"]["verdict"].to_string();
    
    return if verdict.contains("Authentic") { 1 } else { 2 }
}