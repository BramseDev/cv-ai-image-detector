import type { RequestHandler } from '@sveltejs/kit';

export const POST: RequestHandler = async ({ request }) => {
    const formData = await request.formData();
    const file = formData.get('image');

    if (!(file instanceof File)) {
        return new Response(JSON.stringify({ error: 'Invalid file' }), { status: 400 });
    }

    const forwardFormData = new FormData();
    forwardFormData.append('image', file);

    const response = await fetch('http://localhost:8080/upload', {
        method: 'POST',
        body: forwardFormData
    });

    if (!response.ok) {
        return new Response(JSON.stringify({ error: 'Failed to analyze image' }), {
            status: 500
        });
    }

    const result = await response.json();

    // Handle new JSON structure with "analysis" section
    const analysis = result?.analysis;

    if (!analysis || typeof analysis.probability !== 'number') {
        console.error("Invalid analysis data:", analysis);
        return new Response(JSON.stringify({ error: 'Invalid response format' }), {
            status: 502
        });
    }

    const responseData = {
        score: analysis.probability,
        verdict: analysis.verdict,
        summary: analysis.summary,
        confidence: analysis.confidence * 100,
        reasoning: analysis.reasoning,
        analysis_quality: analysis.analysis_quality,
        timestamp: result.timestamp,
        detailed_scores: analysis.scores,
        lighting_analysis: result["lighting-analysis"]?.data?.lighting_analysis,
        object_analysis: result["object-coherence"]?.data?.object_analysis,
        color_balance: result["color-balance"]?.data,
        artifact_summary: result.artifacts?.data?.overall_assessment?.artifact_summary
    };

    return new Response(JSON.stringify(responseData), {
        status: 200,
        headers: { 'Content-Type': 'application/json' }
    });
};
