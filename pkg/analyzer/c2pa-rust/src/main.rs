mod claimdata;
mod report;
mod validation;
use std::io::Error;

use report::*;

fn main() -> Result<(), Error> {
    let args: Vec<String> = std::env::args().collect();
    match args.len() {
        1 => {
            return Err(Error::new(std::io::ErrorKind::InvalidInput, "Specify a path"));
        },
        2 => {
            let path = std::path::PathBuf::from(&args[1]);
            let report = Report::from_file(path);
            let json = match serde_json::to_string(&report) { 
                Ok(j) => j,
                Err(_) => String::from("{}")
            };
            println!("{}", json);
        },
        _ => {
            return Err(Error::new(std::io::ErrorKind::InvalidInput, "Too many arguments"));
        }
    };
    Ok(())
}