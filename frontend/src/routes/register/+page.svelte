<script lang="ts">
    import RegisterForm from "$lib/components/auth/register-form.svelte";
    import BackToHome from "$lib/components/auth/back-to-home.svelte";
</script>

<div class="flex min-h-svh w-full flex-col items-center justify-center bg-muted p-6 md:p-10">
    <div class="w-full max-w-sm md:max-w-3xl">
        <BackToHome />
        <RegisterForm />
    </div>
</div>
