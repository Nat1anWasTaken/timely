<script lang="ts">
    import Button from "$lib/components/ui/button/button.svelte";
    import Card from "$lib/components/ui/card/card.svelte";
    import { Github, Heart, Calendar } from "@lucide/svelte";
</script>

<div class="min-h-screen bg-background flex flex-col">
    <!-- Hero Section -->
    <main class="flex-1 flex items-center justify-center px-4 py-16">
        <div class="max-w-4xl mx-auto text-center space-y-8">
            <!-- App Icon & Title -->
            <div class="space-y-4">
                <div class="inline-flex items-center justify-center w-16 h-16 rounded-xl bg-primary text-primary-foreground">
                    <Calendar class="w-8 h-8" />
                </div>
                <h1 class="text-4xl md:text-6xl font-bold tracking-tight">
                    Timely
                </h1>
                <p class="text-xl md:text-2xl text-muted-foreground max-w-2xl mx-auto">
                    A simple, elegant calendar application for sharing your time
                </p>
            </div>

            <!-- Open Source Badge -->
            <Card class="max-w-md mx-auto p-6">
                <div class="text-center space-y-2">
                    <h2 class="text-lg font-semibold">🎉 Free & Open Source</h2>
                    <p class="text-sm text-muted-foreground">
                        Timely is completely free to use and open source. 
                        No subscriptions, no hidden costs.
                    </p>
                </div>
            </Card>

            <!-- Action Buttons -->
            <div class="flex flex-col sm:flex-row gap-4 justify-center items-center">
                <Button 
                    href="https://github.com/Nat1anWasTaken/timely" 
                    variant="default" 
                    size="lg"
                    class="w-full sm:w-auto"
                >
                    <Github class="w-5 h-5" />
                    View on GitHub
                </Button>
                
                <Button 
                    href="#" 
                    variant="outline" 
                    size="lg"
                    class="w-full sm:w-auto"
                >
                    <Heart class="w-5 h-5" />
                    Support the Project
                </Button>
            </div>

            <!-- Footer -->
            <div class="pt-8 text-sm text-muted-foreground">
                <p>Made with ❤️ by the open source community</p>
            </div>
        </div>
    </main>
</div>
