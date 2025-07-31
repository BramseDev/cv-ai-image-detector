use c2pa::{validation_results::StatusCodes, validation_status::ValidationStatus, ValidationResults, ValidationState};
use c2pa_status_tracker::LogKind;
use serde::Serialize;

#[derive(Serialize)]
pub struct ValidationData {
    pub state: ValidationState,
    pub certs_count: usize,
    pub certs_valid: usize,
    pub certs: Vec<Certificate>
}

impl ValidationData {
    pub fn new(
        state: ValidationState,
        certs_count: usize,
        certs_valid: usize,
        certs: Vec<Certificate>
    ) -> ValidationData {
        ValidationData { state, certs_count, certs_valid, certs }
    }
    
    pub fn from_result(result: &ValidationResults) -> ValidationData {
        let state = result.validation_state();
        let (certs, certs_count, certs_valid) = match result.active_manifest() {
            Some(codes) => Certificate::vec_from_codes(codes.clone()),
            None => (Vec::new(), 0, 0)
        };
        ValidationData::new(state, certs_count, certs_valid, certs)
    }
}

#[derive(Serialize)]
pub struct Certificate {
    pub cert_id: String,
    pub cert_code: String,
    pub cert_explanation: String,
    pub cert_valid: bool
}

impl Certificate {
    pub fn new(cert_id: String, cert_code: String, cert_explanation: String, cert_valid: bool) -> Certificate {
        Certificate { cert_id, cert_code, cert_explanation, cert_valid }
    }
    
    pub fn from_status(status: &ValidationStatus) -> Certificate {
        let id = match status.url() {
            Some(url) => url.to_string(),
            None => String::from("n/a")
        };
        let explanation = match status.explanation() {
            Some(expl) => expl.to_string(),
            None => String::from("n/a")
        };
        let is_valid = match status.kind() {
            LogKind::Success => true,
            _ => false
        };
        Certificate::new(id, status.code().to_string(), explanation, is_valid)
    }
    
    pub fn vec_from_codes(codes: StatusCodes) -> (Vec<Certificate>, usize, usize) {
        let mut vector: Vec<Certificate> = Vec::new();
        
        codes.success().iter().for_each(|code| {
            vector.push(Certificate::from_status(code));
        });
        codes.informational().iter().for_each(|code| {
            vector.push(Certificate::from_status(code));
        });
        codes.failure().iter().for_each(|code| {
            vector.push(Certificate::from_status(code));
        });
        
        let len = &vector.iter().count();
        (vector, *len, codes.success().iter().count())
    }
}