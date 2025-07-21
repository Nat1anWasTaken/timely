<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import {
        Card,
        CardContent,
        CardDescription,
        CardHeader,
        CardTitle
    } from "$lib/components/ui/card";
    import { Trash2 } from "@lucide/svelte";
</script>

<Card class="border-destructive">
    <CardHeader>
        <CardTitle class="flex items-center gap-2 text-destructive">
            <Trash2 class="h-5 w-5" />
            Danger Zone
        </CardTitle>
        <CardDescription>Permanently delete your account and all data</CardDescription>
    </CardHeader>
    <CardContent>
        <Button variant="destructive" size="sm">Delete Account</Button>
    </CardContent>
</Card>
