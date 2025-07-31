use std::{fs::File, io::Error, path::PathBuf};
use c2pa::{format_from_path, Reader, ValidationState};
use serde::Serialize;

use crate::{claimdata::ClaimData, validation::ValidationData};

#[derive(serde::Serialize)]
pub struct Report {
    file_name: String,
    file_type: String,
    verdict: Verdict,
    score: u8,
    score_confidence: u8,
    claims_found: bool,
    claims_count: usize,
    claims: Vec<ClaimData>,
    validation: ValidationData
}

impl Report {
    pub fn new(
        file_name: String,
        file_type: String,
        verdict: Verdict,
        score: u8,
        score_confidence: u8,
        claims_found: bool,
        claims_count: usize,
        claims: Vec<ClaimData>,
        validation: ValidationData
    ) -> Report {
        Report { file_name, file_type, verdict, score, score_confidence, claims_found, claims_count, claims, validation }
    }
    
    pub fn from_file(path: PathBuf) -> Report {
        let file_name = match path.file_name() {
            Some(n) => String::from(n.to_str().unwrap()),
            None => String::from("n/a")
        };
        let file_type = file_name.split(".")
            .last()
            .unwrap()
            .to_string();
        let (claims, validation_data) = handle_file(path);
        let mut score = 0_u8;
        let mut score_confidence = 0_u8;
        let mut claims_found = false;
        let iterator = claims.iter();
        let claims_count = iterator.clone().count();
        if claims_count != 0 {
            score = 1_u8;
            score_confidence = 1_u8;
            claims_found = true;
            let suspicious_generators = [
                "chatgpt",
                "gpt",
                "gpt-3",
                "gpt-4",
                "gpt-4o",
                "microsoft responsible ai image provenance",
                "midjourney",
                "stable diffusion",
                "adobe firefly",
                "leonardo",
                "dall-e"
            ]; // TODO: test this
            let manipulation_generators = ["photoshop", "gimp"]; //TODO: see above
            iterator.for_each(|claim| {
                claim.claim_generator.iter().for_each(|generator| {
                    if suspicious_generators.contains(&generator.to_lowercase().as_str()) {
                        score += 100_u8;
                        score_confidence += 50_u8;
                    } else if manipulation_generators.contains(&generator.to_lowercase().as_str()) {
                        score += 50_u8;
                        score_confidence += 50_u8;
                    }
                });
            });    
        };
        if validation_data.certs_count != 0 {
            score += 20_u8;
            score_confidence += 20_u8;
            match validation_data.state {
                ValidationState::Valid => {
                    score_confidence += 40_u8;
                },
                ValidationState::Trusted => {
                    score_confidence += 60_u8;
                }
                ValidationState::Invalid => {
                    score += 60_u8;
                    score_confidence += 20_u8;
                }
            }
        }
        if score > 100 { score = 100 };
        if score_confidence > 100 { score_confidence = 100 };
        let verdict = Verdict::from_score(score, score_confidence);
        Report::new(file_name, file_type, verdict, score, score_confidence, claims_found, claims_count, claims, validation_data)
    }
}

#[derive(Serialize)]
pub enum Verdict {
    Generated,
    Modified,
    Genuine,
    Unknown,
}

impl Verdict {
    pub fn from_score(score: u8, score_confidence: u8) -> Verdict {
        if score == 0 && score_confidence == 0 {
            return Verdict::Unknown
        } else if score < 21 {
            if score_confidence > 40 { return Verdict::Genuine } else { return Verdict::Modified }
        } else if score < 81 {
            return Verdict::Modified
        } else {
            return Verdict::Generated
        }
    }
}

fn read_c2pa(file: File, path: PathBuf) -> Result<(Vec<ClaimData>, ValidationData), Error> {
    let format = format_from_path(&path).unwrap();
    match Reader::from_stream(&format, &file) {
        Ok(reader) => {
            //println!("c2pa block found");
            let data = ClaimData::vec_from_manifest(reader.manifests());
            let validation_data = match reader.validation_results() {
                Some(res) => ValidationData::from_result(res),
                None => ValidationData::new(ValidationState::Invalid, 0, 0, Vec::new())
            };
            return Ok((data, validation_data));
        }
        Err(c2pa::Error::JumbfNotFound) => {
           //println!("no data");
           return Err(Error::new(std::io::ErrorKind::NotFound, "No data found"));
        },
        Err(_) => {
            //println!("other error");
            return Err(Error::new(std::io::ErrorKind::Other, "Other error"));
        }
    };
}

fn handle_file(path: std::path::PathBuf) -> (Vec<ClaimData>, ValidationData) {
    match File::open(&path) {
        Ok(f) => {
            match read_c2pa(f, path) {
                Ok(data) => data,
                Err(_) => (Vec::new(), ValidationData::new(ValidationState::Invalid, 0, 0, Vec::new()))
            }
        },
        Err(_) => {
            //println!("foiled");
            (Vec::new(), ValidationData::new(ValidationState::Invalid, 0, 0, Vec::new()))
        }
    }
}