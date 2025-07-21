<script lang="ts">
    import "../app.css";
    import { page } from "$app/state";
    import Navbar from "$lib/components/dashboard/navbar.svelte";
    import { ModeWatcher } from "mode-watcher";

    let { children } = $props();

    let hideNavbar = $derived(["/login", "/register"].includes(page.url.pathname));
</script>

<ModeWatcher />

<div class="flex h-screen w-screen flex-col">
    {#if !hideNavbar}
        <Navbar />
    {/if}

    <div class="flex h-full w-full flex-1 items-center justify-center overflow-y-auto">
        {@render children()}
    </div>
</div>
