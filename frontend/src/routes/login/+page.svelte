<script lang="ts">
    import LoginForm from "$lib/components/auth/login-form.svelte";
    import BackToHome from "$lib/components/auth/back-to-home.svelte";
    import { createMutation } from "@tanstack/svelte-query";
    import { api } from "$lib/api";
    import { goto } from "$app/navigation";
    import { toast } from "svelte-sonner";
    import type { LoginRequest } from "$lib/types/api";

    const loginMutation = createMutation({
        mutationFn: (credentials: LoginRequest) => api.login(credentials),
        onSuccess: () => {
            goto("/");
        },
        onError: (error) => {
            console.error("Login failed:", error);
            toast.error("Login failed. Please check your credentials and try again.", {
                description: error.message || "An unexpected error occurred."
            });
        }
    });

    function handleSubmit(email: string, password: string) {
        $loginMutation.mutate({ email, password });
    }
</script>

<div class="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
    <div class="min-w-sm">
        <BackToHome />
        <LoginForm onSubmit={handleSubmit} />
    </div>
</div>
