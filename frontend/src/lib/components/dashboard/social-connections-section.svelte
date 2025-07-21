<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import {
        Card,
        CardContent,
        CardDescription,
        CardHeader,
        CardTitle
    } from "$lib/components/ui/card";
    import { Link, Chrome } from "@lucide/svelte";
</script>

<Card>
    <CardHeader>
        <CardTitle class="flex items-center gap-2">
            <Link class="h-5 w-5" />
            Social Connections
        </CardTitle>
        <CardDescription>Manage your connected social accounts</CardDescription>
    </CardHeader>
    <CardContent>
        <div class="space-y-3">
            <div class="flex items-center justify-between rounded-lg border p-3">
                <div class="flex items-center space-x-3">
                    <Chrome class="h-5 w-5 text-blue-600" />
                    <div>
                        <p class="text-sm font-medium">Google</p>
                        <p class="text-xs text-muted-foreground">Connected</p>
                    </div>
                </div>
                <Button variant="outline" size="sm">Disconnect</Button>
            </div>
        </div>
    </CardContent>
</Card>
