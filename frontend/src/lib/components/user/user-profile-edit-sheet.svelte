<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import {
        Sheet,
        SheetContent,
        SheetDescription,
        SheetHeader,
        SheetTitle,
        SheetTrigger
    } from "$lib/components/ui/sheet";
    import type { User } from "$lib/types/api";
    import UserProfileEditForm from "./user-profile-edit-form.svelte";

    interface Props {
        user: User;
        onUserUpdate: (updatedUser: User) => void;
        children: import("svelte").Snippet;
    }

    let { user, onUserUpdate, children }: Props = $props();

    let open = $state(false);

    function handleSuccess(updatedUser: User) {
        onUserUpdate(updatedUser);
        open = false;
    }

    function handleCancel() {
        open = false;
    }
</script>

<Sheet bind:open>
    <SheetTrigger>
        {@render children()}
    </SheetTrigger>
    <SheetContent class="sm:max-w-lg">
        <SheetHeader>
            <SheetTitle>Edit Profile</SheetTitle>
            <SheetDescription>
                Update your username and display name. Your username is used in your public calendar URL.
            </SheetDescription>
        </SheetHeader>

        <UserProfileEditForm 
            {user} 
            onSuccess={handleSuccess} 
            onCancel={handleCancel} 
        />
    </SheetContent>
</Sheet>