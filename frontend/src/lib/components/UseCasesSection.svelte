<script>
    const useCases = [
        {
            title: "Mis- & Disinformation",
            description:
                "Detect AI-generated misinformation and deepfakes in media.",
            svg: "M280-280h280v-80H280v80Zm0-160h400v-80H280v80Zm0-160h400v-80H280v80Zm-80 480q-33 0-56.5-23.5T120-200v-560q0-33 23.5-56.5T200-840h560q33 0 56.5 23.5T840-760v560q0 33-23.5 56.5T760-120H200Zm0-80h560v-560H200v560Zm0-560v560-560Z",
        },
        {
            title: "Fake News",
            description:
                "Stop fake news from spreading to harm democratic processes.",
            svg: "M160-120q-33 0-56.5-23.5T80-200v-640l67 67 66-67 67 67 67-67 66 67 67-67 67 67 66-67 67 67 67-67 66 67 67-67v640q0 33-23.5 56.5T800-120H160Zm0-80h280v-240H160v240Zm360 0h280v-80H520v80Zm0-160h280v-80H520v80ZM160-520h640v-120H160v120Z",
        },
        {
            title: "Fake photo evidence",
            description:
                "Distinguish fraudulent accident reports with AI-generated images.",
            svg: "M200-120q-33 0-56.5-23.5T120-200v-560q0-33 23.5-56.5T200-840h560q33 0 56.5 23.5T840-760v560q0 33-23.5 56.5T760-120H200Zm0-80h560v-560H200v560Zm40-80h480L570-480 450-320l-90-120-120 160Zm-40 80v-560 560Z",
        },
    ];
</script>

<section class="w-full bg-white py-20">
    <div class="max-w-6xl mx-auto px-4">
        <h2 class="text-3xl font-semibold text-center text-gray-800 mb-16">
            Why Should You Use an AI Image Detector?
        </h2>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {#each useCases as useCase}
                <div class="flex items-start gap-4">
                    <div class="p-4 bg-primary text-white rounded-full">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            height="28px"
                            viewBox="0 -960 960 960"
                            width="28px"
                            fill="#e3e3e3"><path d={useCase.svg} /></svg
                        >
                    </div>
                    <div>
                        <h3 class="text-xl font-semibold text-gray-700 mb-1">
                            {useCase.title}
                        </h3>
                        <p class="text-gray-600">
                            {useCase.description}
                        </p>
                    </div>
                </div>
            {/each}
        </div>
    </div>
</section>
