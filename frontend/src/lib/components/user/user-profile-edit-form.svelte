<script lang="ts">
    import { Button } from "$lib/components/ui/button/index.js";
    import { Input } from "$lib/components/ui/input/index.js";
    import { Label } from "$lib/components/ui/label/index.js";
    import { api } from "$lib/api.js";
    import { toast } from "svelte-sonner";
    import type { User, UpdateUserProfileRequest } from "$lib/types/api";

    interface Props {
        user: User;
        onSuccess: (updatedUser: User) => void;
        onCancel: () => void;
    }

    let { user, onSuccess, onCancel }: Props = $props();

    // Form state
    let username = $state(user.username);
    let displayName = $state(user.display_name);
    let loading = $state(false);
    let errors = $state<Record<string, string>>({});

    // Username validation function matching Instagram rules
    function isValidUsername(username: string): boolean {
        // Check for consecutive dots
        if (username.includes("..")) {
            return false;
        }

        // Check if starts or ends with dot
        if (username.startsWith(".") || username.endsWith(".")) {
            return false;
        }

        // Check if only contains valid characters: A-Z, a-z, 0-9, _, .
        return /^[A-Za-z0-9_.]+$/.test(username);
    }

    // Validation
    function validateForm() {
        const newErrors: Record<string, string> = {};

        if (!username.trim()) {
            newErrors.username = "Username is required";
        } else if (username.length < 3) {
            newErrors.username = "Username must be at least 3 characters";
        } else if (!isValidUsername(username)) {
            newErrors.username =
                "Username can only contain letters, numbers, underscores, and dots. Cannot start or end with dot, and cannot have consecutive dots";
        }

        if (!displayName.trim()) {
            newErrors.displayName = "Display name is required";
        } else if (displayName.length > 100) {
            newErrors.displayName = "Display name must be less than 100 characters";
        }

        errors = newErrors;
        return Object.keys(newErrors).length === 0;
    }

    // Check if form has changes
    let hasChanges = $derived(username !== user.username || displayName !== user.display_name);

    async function handleSubmit() {
        if (!validateForm()) {
            return;
        }

        if (!hasChanges) {
            toast.info("No changes to save");
            return;
        }

        loading = true;
        errors = {};

        try {
            const updateRequest: UpdateUserProfileRequest = {};

            if (username !== user.username) {
                updateRequest.username = username;
            }

            if (displayName !== user.display_name) {
                updateRequest.display_name = displayName;
            }

            const response = await api.updateUserProfile(updateRequest);

            if (response.success) {
                toast.success(response.message || "Profile updated successfully");
                onSuccess(response.user);
            } else {
                toast.error(response.message || "Failed to update profile");
            }
        } catch (err: any) {
            if (err.status === 400 && err.data?.error) {
                // Handle specific validation errors from backend
                if (err.data.error === "username_taken") {
                    errors = { username: "Username is already taken" };
                } else if (
                    err.data.error === "invalid_username" ||
                    err.data.error === "invalid_username_format"
                ) {
                    errors = { username: err.data.message || "Invalid username" };
                } else if (err.data.error === "invalid_display_name") {
                    errors = { displayName: err.message || "Invalid display name" };
                } else {
                    toast.error(err.message || "Failed to update profile");
                }
            } else {
                toast.error(err.message || "An error occurred while updating your profile");
            }
        } finally {
            loading = false;
        }
    }

    function handleUsernameInput(e: Event) {
        const target = e.currentTarget as HTMLInputElement;
        username = target.value;
        // Clear username error when user starts typing
        if (errors.username) {
            errors = { ...errors, username: "" };
        }
    }

    function handleDisplayNameInput(e: Event) {
        const target = e.currentTarget as HTMLInputElement;
        displayName = target.value;
        // Clear display name error when user starts typing
        if (errors.displayName) {
            errors = { ...errors, displayName: "" };
        }
    }
</script>

<div class="space-y-4">
    <div class="space-y-4 p-4">
        <div class="space-y-2">
            <Label for="username" class="text-sm font-medium">Username</Label>
            <Input
                id="username"
                type="text"
                placeholder="Enter username"
                value={username}
                oninput={handleUsernameInput}
                class={errors.username ? "border-destructive" : ""}
                disabled={loading}
            />
            {#if errors.username}
                <p class="text-destructive text-sm">{errors.username}</p>
            {/if}
            <p class="text-muted-foreground text-xs">
                Your username is how others can find your public calendar at /{username}
            </p>
        </div>

        <div class="space-y-2">
            <Label for="display-name" class="text-sm font-medium">Display Name</Label>
            <Input
                id="display-name"
                type="text"
                placeholder="Enter display name"
                value={displayName}
                oninput={handleDisplayNameInput}
                class={errors.displayName ? "border-destructive" : ""}
                disabled={loading}
            />
            {#if errors.displayName}
                <p class="text-destructive text-sm">{errors.displayName}</p>
            {/if}
            <p class="text-muted-foreground text-xs">
                This is the name that will be displayed on your profile and calendar
            </p>
        </div>
    </div>

    <div class="flex justify-end space-x-2 border-t p-4">
        <Button variant="outline" onclick={onCancel} disabled={loading}>Cancel</Button>
        <Button onclick={handleSubmit} disabled={loading || !hasChanges}>
            {#if loading}
                <div
                    class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
                ></div>
            {/if}
            Save Changes
        </Button>
    </div>
</div>
