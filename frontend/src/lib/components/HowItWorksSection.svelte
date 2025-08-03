<script>
    const cards = [
        {
            title: "Upload your image",
            description:
                "Drag and drop or upload an image. We support JPEG, PNG, and WebP formats.",
        },
        {
            title: "AI analysis starts",
            description:
                "Our AI models detect patterns of manipulation, deepfake signs, or synthetic generation.",
        },
        {
            title: "Get your results",
            description:
                "View a detailed analysis report with an authenticity score and key findings.",
        },
    ];
</script>

<section class="w-full bg-gray-50 py-20">
    <div class="max-w-6xl mx-auto px-4">
        <h2 class="text-3xl font-semibold text-center text-gray-800 mb-12">
            How It Works
        </h2>
        <ul class="steps steps-horizontal w-full">
            <li class="step step-primary">Upload Image</li>
            <li class="step">AI Analysis</li>
            <li class="step">Get Results</li>
        </ul>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mt-12">
            {#each cards as card}
                <div class="card bg-base-100 shadow-md">
                    <div class="card-body">
                        <h2 class="card-title">{card.title}</h2>
                        <p>{card.description}</p>
                    </div>
                </div>
            {/each}
        </div>
    </div>
</section>
