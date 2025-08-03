<script lang="ts">
    let { fileSelected } = $props();
    let selectedFile: File | null = $state(null);

    function handleDrop(event) {
        event.preventDefault();

        const files = event.dataTransfer
            ? event.dataTransfer.files
            : event.target.files;
        if (files.length > 0) {
            const file = files[0];
            const allowedTypes = ["image/jpeg", "image/png", "image/webp"];
            if (!allowedTypes.includes(file.type)) {
                alert("Only JPEG, PNG, and WebP formats are supported.");
                return;
            }
            if (file.size > 10 * 1024 * 1024) {
                alert("File size must be less than 10 MB.");
                return;
            }
            selectedFile = file;
            console.log("File selected:", file.name);
            fileSelected({ file: selectedFile });
        }
    }

    function handleDragOver(event) {
        event.preventDefault();
    }
</script>

<div class="w-full h-full">
    <div
        role="form"
        ondrop={handleDrop}
        ondragover={handleDragOver}
        class="relative w-full h-full border-2 border-dashed rounded-xl transition-all duration-300 flex flex-col items-center justify-center p-6 text-center gap-4 cursor-pointer
                border-base-300 bg-base-200/50 hover:border-primary hover:bg-primary/5"
    >
        <div class="text-5xl text-base-content/60">
            <svg
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                stroke-width="1.5"
                stroke="currentColor"
                class="w-12 h-12 mx-auto"
            >
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5m-13.5-9L12 3m0 0 4.5 4.5M12 3v13.5"
                />
            </svg>
        </div>
        <div class="space-y-2">
            {#if selectedFile}
                <p class="text-base-content font-medium">
                    <span class="text-primary">{selectedFile.name}</span> selected
                </p>
            {:else}
                <p class="font-medium text-base-content">
                    <span class="text-primary">Click to upload</span> or drag an
                    drop an image
                </p>
            {/if}
            <p class="text-sm text-base-content/70">
                Supports JPEG, PNG, and WebP formats. Max. 10 MB
            </p>
        </div>
        <input
            type="file"
            accept="image/jpeg,image/png,image/webp"
            class="absolute inset-0 opacity-0 cursor-pointer"
            onchange={handleDrop}
        />
    </div>
</div>
