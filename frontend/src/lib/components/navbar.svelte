<script>
    import { Card, CardTitle } from "$lib/components/ui/card/index.ts";
    import { Button } from "$lib/components/ui/button/index.ts";
    import ModeToggle from "$lib/components/dashboard/mode-toggle.svelte";
</script>

<div class="m-4 flex max-h-16">
    <Card class="w-full flex-row items-center justify-between p-4">
        <CardTitle>
            <a class="text-xl" href="/">Timely</a>
        </CardTitle>
        <div class="flex items-center">
            <ModeToggle />
            <Button variant="outline" class="ml-4" href="/login">Login</Button>
        </div>
    </Card>
</div>
<x></x>
