<script lang="ts">
    import { page } from "$app/state";
    import Navbar from "$lib/components/navbar/navbar.svelte";
    import { Toaster } from "$lib/components/ui/sonner";
    import { QueryClient, QueryClientProvider } from "@tanstack/svelte-query";
    import { SvelteQueryDevtools } from "@tanstack/svelte-query-devtools";
    import "../app.css";

    let { children } = $props();

    let hideNavbar = $derived(["/login", "/register"].includes(page.url.pathname));

    const queryClient = new QueryClient();
</script>

<QueryClientProvider client={queryClient}>
    <SvelteQueryDevtools />
    <Toaster />
    <div class="flex h-screen w-screen flex-col">
        {#if !hideNavbar}
            <Navbar />
        {/if}

        <div class="flex h-full w-full flex-1 items-center justify-center overflow-y-auto">
            {@render children()}
        </div>
    </div>
</QueryClientProvider>
