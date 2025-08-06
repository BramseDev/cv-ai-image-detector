use serde::Serialize;

#[derive(Serialize)]
pub struct EvalResult {
    pub expected_result: u8,
    pub actual_result: u8,
    pub file_name: String
}

impl EvalResult {
    pub fn new(expected_result: u8, actual_result: u8, file_name: String) -> EvalResult {
        EvalResult { expected_result, actual_result, file_name }
    }
}

pub trait Stringify {
    fn to_string(&self) -> String;
}

impl Stringify for EvalResult {
    fn to_string(&self) -> String {
        format!("{}\t{}\t{}", self.expected_result, self.actual_result, self.file_name)
    }
}

#[derive(Serialize)]
pub struct EvalReport {
    pub files_analyzed: usize,
    pub expected_result: u8,
    pub hits: usize,
    pub misses: usize,
    pub fails: usize,
    pub accuracy: f32,
    pub results: Vec<EvalResult>
}

impl EvalReport {
    pub fn from(results: Vec<EvalResult>) -> EvalReport {
        let files_analyzed = results.iter().count();
        if files_analyzed == 0 {
            return EvalReport {
                files_analyzed,
                expected_result: 0,
                hits: 0,
                misses: 0,
                fails: 0,
                accuracy: 0.0,
                results: results
            }
        }
        
        let expected_result = results.first().unwrap().expected_result;
        let mut hits: usize = 0;
        let mut misses: usize = 0;
        let mut fails: usize = 0;
        results.iter().for_each(|result| {
            if result.actual_result == 0 {
                fails += 1;
            } else if result.actual_result == result.expected_result {
                hits += 1;
            } else {
                misses += 1;
            }
        });
        
        let accuracy: f32 = hits as f32 / files_analyzed as f32;
        
        EvalReport { files_analyzed, expected_result, hits, misses, fails, accuracy, results }
    }
}