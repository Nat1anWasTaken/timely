<script>
    import { Card, CardTitle } from "$lib/components/ui/card/index.js";
    import { Button } from "$lib/components/ui/button/index.js";
    import ModeToggle from "$lib/components/mode-toggle.svelte";
</script>

<div class="m-4 flex max-h-16">
    <Card class="w-full flex-row items-center justify-between p-4">
        <CardTitle class="text-xl">Timely</CardTitle>
        <div class="flex items-center">
            <ModeToggle />
            <Button variant="outline" class="ml-4" href="/login">Login</Button>
        </div>
    </Card>
</div>
<x></x>
