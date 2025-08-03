<script lang="ts">
    import { page } from "$app/state";

    $: user = page.data.user;
</script>

{#snippet links()}
    <li>
        <a href="/#how-it-works">How It Works</a>
    </li>
    <li>
        <a href="/#use-cases">Use Cases</a>
    </li>
    <li>
        <a href="/#faq">FAQ</a>
    </li>
{/snippet}

{#snippet accountLinks()}
    {#if user}
        <li><a href="/history">My History</a></li>
        <li>
            <form method="POST" action="?/logout">
                <button type="submit">Logout</button>
            </form>
        </li>
    {:else}
        <li><a href="/login">Login</a></li>
    {/if}
{/snippet}

<div class="drawer">
    <input id="nav-drawer" type="checkbox" class="drawer-toggle" />
    <div class="drawer-content">
        <nav class="navbar bg-base-100 max-w-11/12 mx-auto">
            <div class="flex-1">
                <a href="/frontend/static" class="text-xl font-medium px-4 py-2"
                    >AI Image Detector</a
                >
            </div>

            <div class="flex-none lg:hidden">
                <label for="nav-drawer" class="btn btn-square btn-ghost">
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke-width="1.5"
                        stroke="currentColor"
                        class="w-6 h-6"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5"
                        />
                    </svg>
                </label>
            </div>

            <div class="flex-none hidden lg:block">
                <ul class="menu menu-horizontal items-center space-x-2 px-1">
                    {@render links()}
                    <div class="pl-2">
                        <button
                            class="btn btn-circle"
                            popovertarget="popover-1"
                            style="anchor-name:--anchor-1"
                            aria-label="Avater"
                        >
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                height="24px"
                                viewBox="0 -960 960 960"
                                width="24px"
                                fill="#444444"
                                ><path
                                    d="M480-480q-66 0-113-47t-47-113q0-66 47-113t113-47q66 0 113 47t47 113q0 66-47 113t-113 47ZM160-160v-112q0-34 17.5-62.5T224-378q62-31 126-46.5T480-440q66 0 130 15.5T736-378q29 15 46.5 43.5T800-272v112H160Zm80-80h480v-32q0-11-5.5-20T700-306q-54-27-109-40.5T480-360q-56 0-111 13.5T260-306q-9 5-14.5 14t-5.5 20v32Zm240-320q33 0 56.5-23.5T560-640q0-33-23.5-56.5T480-720q-33 0-56.5 23.5T400-640q0 33 23.5 56.5T480-560Zm0-80Zm0 400Z"
                                /></svg
                            >
                        </button>

                        <ul
                            class="dropdown dropdown-center menu mt-3 w-36 rounded-box bg-base-100 shadow-sm"
                            popover
                            id="popover-1"
                            style="position-anchor:--anchor-1"
                        >
                            {@render accountLinks()}
                        </ul>
                    </div>
                </ul>
            </div>
        </nav>
    </div>

    <div class="drawer-side z-10">
        <label for="nav-drawer" class="drawer-overlay"></label>
        <ul class="menu p-4 w-80 h-full bg-base-100 space-y-3.5">
            {@render links()}
            {@render accountLinks()}
        </ul>
    </div>
</div>
