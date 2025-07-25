<script lang="ts">
    import { api } from "$lib/api";
    import { Button } from "$lib/components/ui/button";
    import * as Card from "$lib/components/ui/card";
    import { Input } from "$lib/components/ui/input";
    import { page } from "$app/state";
    import { Label } from "$lib/components/ui/label";

    interface Props {
        onSubmit: (
            username: string,
            displayName: string,
            email: string,
            password: string,
            confirmPassword: string
        ) => void;
    }

    const id = $props.id();
    const { onSubmit }: Props = $props();

    let username = $state("");
    let displayName = $state("");
    let email = $state("");
    let password = $state("");
    let confirmPassword = $state("");

    const passwordsMatch = $derived(password === confirmPassword && password.length > 0);
    const showError = $derived(confirmPassword.length > 0 && !passwordsMatch);

    let googleAuthUrl = $derived(
        `${api.baseUrl}/api/auth/google/login?mode=login&from=${page.url.origin}`
    );
</script>

<Card.Root class="mx-auto w-full max-w-sm">
    <Card.Header>
        <Card.Title class="text-2xl">Register</Card.Title>
        <Card.Description>Fill out the form to create a new account</Card.Description>
    </Card.Header>
    <Card.Content>
        <form
            onsubmit={(event) => {
                event.preventDefault();
                if (passwordsMatch) {
                    onSubmit(username, displayName, email, password, confirmPassword);
                }
            }}
        >
            <div class="grid gap-4">
                <div class="grid gap-2">
                    <Label for="username-{id}">Username</Label>
                    <Input
                        id="username-{id}"
                        type="username"
                        placeholder=""
                        required
                        bind:value={username}
                    />
                </div>
                <div class="grid gap-2">
                    <Label for="display-name-{id}">Display Name</Label>
                    <Input
                        id="display-name-{id}"
                        type="text"
                        placeholder=""
                        required
                        bind:value={displayName}
                    />
                </div>
                <div class="grid gap-2">
                    <Label for="email-{id}">Email</Label>
                    <Input
                        id="email-{id}"
                        type="email"
                        placeholder="m@example.com"
                        required
                        bind:value={email}
                    />
                </div>
                <div class="grid gap-2">
                    <div class="flex items-center">
                        <Label for="password-{id}">Password</Label>
                    </div>
                    <Input id="password-{id}" type="password" required bind:value={password} />
                </div>
                <div class="grid gap-2">
                    <div class="flex items-center">
                        <Label for="password-{id}">Confirm Password</Label>
                    </div>
                    <Input
                        id="password-confirm-{id}"
                        type="password"
                        required
                        bind:value={confirmPassword}
                    />
                    {#if showError}
                        <p class="text-sm text-destructive">Passwords do not match</p>
                    {/if}
                </div>
                <Button
                    type="submit"
                    class="w-full"
                    onclick={() => {
                        if (passwordsMatch) {
                            onSubmit(username, displayName, email, password, confirmPassword);
                        }
                    }}
                    >Register
                </Button>
                <Button variant="outline" class="w-full" href={googleAuthUrl}>
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                        <path
                            d="M12.48 10.92v3.28h7.84c-.24 1.84-.853 3.187-1.787 4.133-1.147 1.147-2.933 2.4-6.053 2.4-4.827 0-8.6-3.893-8.6-8.72s3.773-8.72 8.6-8.72c2.6 0 4.507 1.027 5.907 2.347l2.307-2.307C18.747 1.44 16.133 0 12.48 0 5.867 0 .307 5.387.307 12s5.56 12 12.173 12c3.573 0 6.267-1.173 8.373-3.36 2.16-2.16 2.84-5.213 2.84-7.667 0-.76-.053-1.467-.173-2.053H12.48z"
                            fill="currentColor"
                        />
                    </svg>
                    Continue with Google
                </Button>
            </div>
            <div class="mt-4 text-center text-sm">
                Don't have an account?
                <a href="/login" class="underline"> Sign up </a>
            </div>
        </form>
    </Card.Content>
</Card.Root>
