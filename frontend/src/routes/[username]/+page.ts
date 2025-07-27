import { api } from "$lib/api";
import type { PageLoad } from "./$types";

export const ssr = false;

export const load: PageLoad = async ({ params }) => {
    try {
        // Try to get the authenticated user's profile first
        const currentUser = await api.getUserProfile();

        if (currentUser.user?.username === params.username) {
            // User is viewing their own profile
            return {
                isViewingSelf: true,
                message: "",
                user: currentUser.user,
                publicUser: null
            };
        }
    } catch {
        // User is not authenticated, continue to public profile
    }

    try {
        // Get the public profile for the requested username
        const publicProfile = await api.getPublicUserProfile(params.username);

        return {
            isViewingSelf: false,
            message: "",
            user: null,
            publicUser: publicProfile.user
        };
    } catch {
        return {
            isViewingSelf: false,
            message: "User not found.",
            user: null,
            publicUser: null
        };
    }
};
