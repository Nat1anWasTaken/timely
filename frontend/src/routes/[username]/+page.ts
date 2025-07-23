import { api } from "$lib/api";
import type { PageLoad } from "./$types";

export const ssr = false;

export const load: PageLoad = async ({ params }) => {
    const user = await api.getUserProfile();

    if (user.user?.username === params.username) {
        return {
            isViewingSelf: true,
            message: "",
            user: user
        };
    } else {
        return {
            isViewingSelf: false,
            message: "Viewing other's profile is currently not avaiable."
        };
    }
};
