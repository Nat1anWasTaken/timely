<script lang="ts">
    import { Button } from "$lib/components/ui/button/index.js";
    import { ChevronLeft, ChevronRight } from "@lucide/svelte";

    let currentDate = new Date();
    let { year = $bindable(currentDate.getFullYear()), month = $bindable(currentDate.getMonth()) } =
        $props();

    function updateMonth(newMonth: number) {
        if (newMonth < 0) {
            month = 11;
            year--;
        } else if (newMonth > 11) {
            month = 0;
            year++;
        } else {
            month = newMonth;
        }
    }
</script>

<div class="flex flex-row items-center gap-2">
    <Button variant="ghost" size="icon" onclick={() => updateMonth(month - 1)}>
        <ChevronLeft />
    </Button>
    <span class="w-[10em] text-center text-lg font-semibold">
        {new Intl.DateTimeFormat("en-US", { month: "long" }).format(new Date(year, month))}
        {year}
    </span>
    <Button variant="ghost" size="icon" onclick={() => updateMonth(month + 1)}>
        <ChevronRight />
    </Button>
</div>
