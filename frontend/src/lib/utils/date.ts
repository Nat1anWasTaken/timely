export function getMonthBoundaries(year: number, month: number) {
    const startOfMonth = new Date(year, month, 1);
    const endOfMonth = new Date(year, month + 1, 0);
    
    return {
        start_timestamp: Math.floor(startOfMonth.getTime() / 1000).toString(),
        end_timestamp: Math.floor(endOfMonth.getTime() / 1000).toString()
    };
}

export function createQueryKey(year: number, month: number) {
    return ['calendar-events', year, month];
}