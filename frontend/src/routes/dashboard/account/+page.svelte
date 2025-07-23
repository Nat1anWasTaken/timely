<script lang="ts">
    import ProfileSection from "$lib/components/dashboard/account/profile-section.svelte";
    import PasswordSection from "$lib/components/dashboard/account/password-section.svelte";
    import SocialConnectionsSection from "$lib/components/dashboard/account/social-connections-section.svelte";
    import DangerZoneSection from "$lib/components/dashboard/account/danger-zone-section.svelte";
</script>

<div class="w-2xl max-w-[90vw] space-y-6">
    <div class="hidden md:block">
        <h1 class="text-2xl font-bold tracking-tight">Account Settings</h1>
    </div>

    <ProfileSection />
    <PasswordSection />
    <SocialConnectionsSection />
    <DangerZoneSection />
</div>
