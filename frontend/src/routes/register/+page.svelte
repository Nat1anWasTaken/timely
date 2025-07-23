<script lang="ts">
    import RegisterForm from "$lib/components/auth/register-form.svelte";
    import BackToHome from "$lib/components/auth/back-to-home.svelte";

    import { createMutation } from "@tanstack/svelte-query";

    import { api } from "$lib/api";
    import { toast } from "svelte-sonner";
    import { goto } from "$app/navigation";
    import { createUserDataQuery } from "$lib/globalQueries";

    const userDataQuery = createUserDataQuery();

    const registerMutation = createMutation({
        mutationFn: (data: {
            email: string;
            displayName: string;
            password: string;
            confirmPassword: string;
        }) =>
            api.register({
                username: data.email, // Assuming username is the same as email for registration
                display_name: data.displayName,
                email: data.email,
                password: data.password
            }),
        onSuccess: () => {
            goto("/");
            toast.success("Registration successful!");
            $userDataQuery.refetch();
        },
        onError: (error) => {
            console.error("Registration failed:", error);
            toast.error("Registration failed. Please check your details and try again.", {
                description: error.message || "An unexpected error occurred."
            });
        }
    });

    function handleSubmit(
        email: string,
        displayName: string,
        password: string,
        confirmPassword: string
    ) {
        $registerMutation.mutate({ email, displayName, password, confirmPassword });
    }
</script>

<div class="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
    <div class="min-w-sm">
        <BackToHome />
        <RegisterForm onSubmit={handleSubmit} />
    </div>
</div>
