<script lang="ts">
    import type { User } from "$lib/types/api.ts";
    import * as Avatar from "../ui/avatar/index.ts";

    interface Props {
        user?: User | null;
    }

    let { user }: Props = $props();
</script>

<Avatar.Root>
    <Avatar.Image src={user?.picture} alt={user?.username} />
    <Avatar.Fallback>{user?.username.charAt(0)}</Avatar.Fallback>
</Avatar.Root>
