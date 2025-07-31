use std::collections::HashMap;
use c2pa::Manifest;

#[derive(serde::Serialize)]
pub struct ClaimData {
    pub claim_id: String,
    pub claim_issuer: String,
    pub claim_generator: Vec<String>
}

impl ClaimData {
    pub fn new(claim_id: String, claim_issuer: String, claim_generator: Vec<String>) -> ClaimData {
       ClaimData { claim_id, claim_issuer, claim_generator } 
    }
    
    pub fn from_manifest(manifest: (&String, &Manifest)) -> ClaimData {
        let issuer = match manifest.1.issuer() {
            Some(iss) => iss,
            None => "none".to_string()
        };
        let claim_gen = manifest.1.claim_generator_info.clone();
        let generators: Vec<String> = match claim_gen {
            Some(list) => {
                let mut gen_vec = Vec::new();
                list.iter().for_each(|ci| {gen_vec.push(ci.name.clone());});
                gen_vec
            },
            None => {
                Vec::new()
            }
        };
        ClaimData::new(manifest.0.clone(), issuer, generators)
    }
    
    pub fn vec_from_manifest(manifest: &HashMap<String, Manifest>) -> Vec<ClaimData> {
        let mut vector: Vec<ClaimData> = Vec::new();
        manifest.iter().for_each(|m| {
            vector.push(ClaimData::from_manifest(m));
        });
        vector
    }
}

pub fn print_data(data: &Vec<ClaimData>) {
    data.iter().for_each(|claim| {
        println!("=== claim ===");
        println!("claim\t{}", claim.claim_id);
        println!("issuer\t{}", claim.claim_issuer);
        claim.claim_generator.iter().for_each(|claim_gen| { println!("gen\t{}", claim_gen) });
    });
}