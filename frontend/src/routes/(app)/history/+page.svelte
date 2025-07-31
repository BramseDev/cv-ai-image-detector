<script lang="ts">
	import Footer from "$lib/components/Footer.svelte";
	import Nav from "$lib/components/Nav.svelte";
	import type { ImageHistory } from "$lib/server/db/schema";
	export let data: { history: ImageHistory[] };
</script>

<div class="flex flex-col min-h-screen">
	<Nav />

	<header>
		<h1
			class="text-4xl font-bold text-center mt-10 font-[GeneralSans-Bold]"
		>
			Upload History
		</h1>
		<p class="text-center text-base-content/70 mt-2">
			The upload history of the images you uploaded in the past
		</p>
	</header>

	<main class="flex-grow flex flex-col items-center w-full mx-auto mt-8">
		<div class="container mx-auto py-12 px-4">
			{#if data.history.length === 0}
				<div class="text-center text-gray-500">
					You haven't uploaded any images yet.
				</div>
			{:else}
				<div class="flex flex-wrap justify-center gap-6">
					{#each data.history as item}
						<div class="card lg:card-side shadow w-2/3">
							<figure class="w-full h-80 lg:w-1/3">
								<img
									src={item.imageData}
									alt="Upload thumbnail"
									class="object-cover w-full rounded-t-lg lg:rounded-l-lg lg:rounded-tr-none"
								/>
							</figure>
							<div class="card-body gap-2">
								<h2 class="card-title">
									{item.imageUrl.split("/").pop()}
								</h2>
								<div>
									<p>
										<span class="font-semibold">Score:</span
										>
										{item.resultScore}%
									</p>
									<p>
										<span class="font-semibold"
											>Confidence:</span
										>
										{item.resultConfidence}%
									</p>
									<p>
										<span class="font-semibold">Date:</span>
										{new Date(
											item.createdAt,
										).toLocaleString()}
									</p>
								</div>
								<div
									class="card-actions mt-auto pt-4 justify-end"
								>
									<form method="POST" action="?/deleteImage">
										<input
											type="hidden"
											name="imageId"
											value={item.id}
										/>
										<button
											class="btn btn-lg btn-soft btn-neutral"
											>Delete</button
										>
									</form>
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</main>

	<Footer />
</div>
