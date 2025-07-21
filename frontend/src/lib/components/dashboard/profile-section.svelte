<script lang="ts">
    import {
        Card,
        CardContent,
        CardDescription,
        CardHeader,
        CardTitle
    } from "$lib/components/ui/card";
    import { User } from "@lucide/svelte";
</script>

<Card>
    <CardHeader>
        <CardTitle class="flex items-center gap-2">
            <User class="h-5 w-5" />
            Profile
        </CardTitle>
        <CardDescription>Your basic account information</CardDescription>
    </CardHeader>
    <CardContent class="space-y-4">
        <div class="flex items-center space-x-4">
            <div class="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
                <User class="h-6 w-6" />
            </div>
            <div class="space-y-1">
                <p class="text-sm font-medium">john.doe@example.com</p>
                <p class="text-xs text-muted-foreground">Member since January 2024</p>
            </div>
        </div>
    </CardContent>
</Card>
