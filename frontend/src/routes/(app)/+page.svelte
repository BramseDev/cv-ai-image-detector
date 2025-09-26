<script lang="ts">
    import FaqSection from "$lib/components/FaqSection.svelte";
    import Footer from "$lib/components/Footer.svelte";
    import HowItWorksSection from "$lib/components/HowItWorksSection.svelte";
    import ImageDropZone from "$lib/components/ImageDropZone.svelte";
    import Nav from "$lib/components/Nav.svelte";
    import UseCasesSection from "$lib/components/UseCasesSection.svelte";

    let selectedFile: File | null = $state(null);
    let imageUrl: string | null = $state(null);
    let resultVerdict: string | null = $state(null);
    let resultDetailedScores: Record<string, number> | null = $state(null);
    let resultLighting: {
        ai_lighting_score: number;
        exposure_uniformity: number;
        light_source_consistency: number;
        shadow_direction_consistency: number;
        anomalies: string[];
    } | null = $state(null);

    let resultObject: {
        ai_coherence_score: number;
        edge_consistency: number;
        lighting_coherence: number;
        perspective_consistency: number;
        anomalies: string[];
    } | null = $state(null);

    let resultColor: {
        ai_color_score?: number;
        imbalance_score: number;
        imbalance_indicators: string[];
        basic_stats?: any;
        channel_ratios?: any;
        contrast?: any;
        histogram_features?: any;
        hsv_analysis?: any;
    } | null = $state(null);

    let resultArtifactSummary: Record<string, boolean> | null = $state(null);

    // AI Model specific results
    let aiModelScore: number | null = $state(null);
    let aiModelExplanation: string | null = $state(null);
    let aiModelConfidence: number | null = $state(null);

    let loading: boolean = $state(false);

    function handleFileSelected(event) {
        selectedFile = event.file;
    }

    async function analyzeImage() {
        if (selectedFile) {
            imageUrl = URL.createObjectURL(selectedFile);

            const formData = new FormData();
            formData.append("image", selectedFile);

            loading = true;

            const response = await fetch("/api/getResultScore", {
                method: "POST",
                body: formData,
            });

            const responseText = await response.text();

            let result;
            try {
                result = JSON.parse(responseText);
                console.log("Parsed JSON:", result);
            } catch (error) {
                console.error("JSON parse error:", error);
                console.error("Response was not valid JSON:", responseText);
                loading = false;
                return;
            }

            // Handle both computer vision and AI model results
            // First, set the computer vision results (original format)

            resultVerdict = result.verdict;
            resultLighting = result.lighting_analysis;
            resultObject = result.object_analysis;
            resultColor = result.color_balance;
            resultArtifactSummary = result.artifact_summary;
            resultDetailedScores = result.detailed_scores;

            // Then, handle AI model results separately
            const aiModelScoreValue = result.detailed_scores?.["ai-model"];
            if (aiModelScoreValue !== undefined) {
                // AI model score: 0 = Human, 1 = AI
                // Convert to percentage for display (0 = 0% AI, 1 = 100% AI)
                aiModelScore = aiModelScoreValue * 100;
                aiModelConfidence = 100; // High confidence since ai-model provides binary results
                aiModelExplanation =
                    result["ai-model"]?.explanation ||
                    `Neural network analysis: ${(aiModelScoreValue * 100).toFixed(1)}% AI probability`;
            } else {
                aiModelScore = null;
                aiModelConfidence = null;
                aiModelExplanation = null;
            }

            loading = false;

            // Then send it to your save endpoint
            const saveForm = new FormData();
            saveForm.append("image", selectedFile);
            saveForm.append(
                "result_score",
                aiModelScore!.toFixed(2).toString(),
            );
            saveForm.append(
                "result_confidence",
                aiModelConfidence!.toFixed(2).toString(),
            );

            await fetch("/?/saveImage", {
                method: "POST",
                body: saveForm,
            });
        }
    }
</script>

<Nav />

<header>
    <h1 class="text-4xl font-bold text-center mt-10 font-[GeneralSans-Bold]">
        AI Image Detector
    </h1>
    <p class="text-center text-base-content/70 mt-2">
        Detect AI-generated images with an accuracy score for free
    </p>
</header>

<main class="flex flex-col items-center w-full mx-auto">
    {#if !imageUrl}
        <div class="text-center">
            <div
                class="flex justify-center items-center mt-8 md:mt-22 h-96 w-10/12 md:w-xl lg:w-2xl xl:w-3xl mx-auto"
            >
                <ImageDropZone fileSelected={handleFileSelected} />
            </div>

            <button
                onclick={analyzeImage}
                class="btn btn-neutral btn-wide btn-lg my-12 font-light"
                disabled={!selectedFile}
            >
                Analyze
            </button>
        </div>
    {/if}

    {#if imageUrl}
        <div class="mt-8 w-full max-w-3xl mx-auto">
            <button
                onclick={() => {
                    window.location.href = "/";
                    setTimeout(() => window.location.reload(), 100);
                }}
                class="btn btn-neutral btn-soft font-light btn-lg mr-4"
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    class="w-4 h-4 mr-1"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                >
                    <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M15 19l-7-7 7-7"
                    />
                </svg>
                Back
            </button>

            {#if loading}
                <!-- Loading indicator -->
                <div class="flex flex-col items-center my-10 space-y-4">
                    <p class="text-lg font-medium text-center">
                        Analyzing your image...
                    </p>
                    <span class="loading loading-bars loading-lg text-primary"
                    ></span>
                </div>
            {:else}
                <!-- Title section -->
                <section class="text-center my-12">
                    <h2
                        class="text-5xl font-extrabold mb-4"
                        style="color: {aiModelScore
                            ? `hsl(${120 - (aiModelScore / 100) * 120}, 70%, 45%)`
                            : 'inherit'}"
                    >
                        {aiModelExplanation}
                    </h2>
                    {#if aiModelScore}
                        <p class="text-2xl text-gray-800">
                            {aiModelScore.toFixed(2)}% chance of being
                            AI-generated
                        </p>
                    {/if}
                    {#if aiModelConfidence}
                        <p class="text-lg text-gray-500 mt-1">
                            Confidence level: {aiModelConfidence.toFixed(2)}%
                        </p>
                    {/if}
                </section>

                <!-- Collapsable details -->
                <div
                    class="collapse collapse-arrow border border-base-300 bg-base-100 mt-8"
                >
                    <input type="checkbox" />
                    <div class="collapse-title font-medium text-xl">
                        Show Detailed Analysis
                    </div>
                    <div class="collapse-content text-black">
                        <div class="text-xl font-medium text-primary mb-2">
                            <p class="text-base text-gray-600 mb-4">
                                Traditional vision techniques look for
                                surface-level artifacts: patterns in pixels,
                                color, or compression. But these signals are
                                quickly outdated as generative models advance.
                                Our trained AI system learns from millions of
                                examples, giving it the ability to spot subtle
                                synthetic fingerprints that older approaches
                                miss. You can view the traditional approach
                                below.
                            </p>
                            <p>Analysis Scores:</p>
                            {#if resultDetailedScores && Object.keys(resultDetailedScores).length > 0}
                                <ul
                                    class="list-disc list-inside text-black mt-2 ml-8"
                                >
                                    {#each Object.entries(resultDetailedScores) as [name, score]}
                                        <li>
                                            {name
                                                .replaceAll("-", " ")
                                                .replaceAll("_", " ")}: {(
                                                score * 100
                                            ).toFixed(1)}%
                                        </li>
                                    {/each}
                                </ul>
                            {:else}
                                <p class="text-gray-500 mt-2 ml-8">
                                    No detailed scores available
                                </p>
                            {/if}
                        </div>
                        {#if resultLighting}
                            <div class="text-xl font-medium text-primary mb-2">
                                <p>Lighting Analysis:</p>
                                <ul
                                    class="list-disc list-inside text-black mt-2 ml-8"
                                >
                                    <li>
                                        Lighting Score: {(
                                            resultLighting.ai_lighting_score *
                                            100
                                        ).toFixed(1)}%
                                    </li>
                                    <li>
                                        Exposure Uniformity: {(
                                            resultLighting.exposure_uniformity *
                                            100
                                        ).toFixed(1)}%
                                    </li>
                                    <li>
                                        Light Consistency: {(
                                            resultLighting.light_source_consistency *
                                            100
                                        ).toFixed(1)}%
                                    </li>
                                    <li>
                                        Shadow Direction Consistency: {(
                                            resultLighting.shadow_direction_consistency *
                                            100
                                        ).toFixed(1)}%
                                    </li>
                                    {#each resultLighting.anomalies as anomaly}
                                        <li>Anomaly: {anomaly}</li>
                                    {/each}
                                </ul>
                            </div>
                        {/if}
                        {#if resultObject}
                            <div class="text-xl font-medium text-primary mb-2">
                                <p>Object Coherence:</p>
                                <ul
                                    class="list-disc list-inside text-black mt-2 ml-8"
                                >
                                    <li>
                                        Coherence Score: {(
                                            resultObject.ai_coherence_score *
                                            100
                                        ).toFixed(1)}%
                                    </li>
                                    <li>
                                        Edge Consistency: {(
                                            resultObject.edge_consistency * 100
                                        ).toFixed(1)}%
                                    </li>
                                    <li>
                                        Lighting Coherence: {(
                                            resultObject.lighting_coherence *
                                            100
                                        ).toFixed(1)}%
                                    </li>
                                    <li>
                                        Perspective Consistency: {(
                                            resultObject.perspective_consistency *
                                            100
                                        ).toFixed(1)}%
                                    </li>
                                </ul>
                            </div>
                        {/if}
                        {#if resultColor}
                            <div class="text-xl font-medium text-primary mb-2">
                                <p>Color Balance:</p>
                                <ul
                                    class="list-disc list-inside text-black mt-2 ml-8"
                                >
                                    <li>
                                        AI Color Score: {(
                                            resultColor.ai_color_score * 100
                                        ).toFixed(1)}%
                                    </li>
                                    <li>
                                        Imbalance Score: {(
                                            resultColor.imbalance_score * 100
                                        ).toFixed(1)}%
                                    </li>
                                    {#each resultColor.imbalance_indicators ?? [] as issue}
                                        <li>
                                            Imbalance: {issue.replaceAll(
                                                "_",
                                                " ",
                                            )}
                                        </li>
                                    {/each}
                                </ul>
                            </div>
                        {/if}
                        {#if resultArtifactSummary}
                            <div class="text-xl font-medium text-primary mb-2">
                                <p>Artifact Summary:</p>
                                <ul
                                    class="list-disc list-inside text-black mt-2 ml-8"
                                >
                                    {#each Object.entries(resultArtifactSummary) as [key, val]}
                                        <li>
                                            {key.replaceAll("_", " ")}: {val
                                                ? "Yes"
                                                : "No"}
                                        </li>
                                    {/each}
                                </ul>
                            </div>
                        {/if}
                    </div>
                </div>
            {/if}
        </div>

        <!-- then below that, just the image centered: -->
        <div class="flex justify-center max-h-100 mt-8">
            <img
                src={imageUrl}
                alt="Uploaded"
                class="object-cover w-xl mb-12 rounded-lg border-2 border-base-300 border"
            />
        </div>
    {/if}

    <div class="flex justify-center items-center w-full" id="how-it-works">
        <HowItWorksSection />
    </div>

    <div class="flex justify-center items-center w-full" id="use-cases">
        <UseCasesSection />
    </div>

    <div class="flex justify-center items-center w-full" id="faq">
        <FaqSection />
    </div>
</main>

<Footer />
