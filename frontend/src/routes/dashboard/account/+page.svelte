<script lang="ts">
    import ProfileSection from "$lib/components/dashboard/profile-section.svelte";
    import EmailSection from "$lib/components/dashboard/email-section.svelte";
    import PasswordSection from "$lib/components/dashboard/password-section.svelte";
    import SocialConnectionsSection from "$lib/components/dashboard/social-connections-section.svelte";
    import DangerZoneSection from "$lib/components/dashboard/danger-zone-section.svelte";
</script>

<div class="w-2xl max-w-[90vw] space-y-6">
    <div class="hidden md:block">
        <h1 class="text-2xl font-bold tracking-tight">Account Settings</h1>
    </div>

    <ProfileSection />
    <EmailSection />
    <PasswordSection />
    <SocialConnectionsSection />
    <DangerZoneSection />
</div>
