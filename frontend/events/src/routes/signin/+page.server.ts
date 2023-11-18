import { redirect } from "@sveltejs/kit";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async () => {
	return {
		greeting: "hi"
	}
}

export const actions: Actions = {
	async signin() {
		const loginUrl = 'http://localhost:4000/signin?redirect=http://localhost:5173/';
		throw redirect(307, loginUrl)
	}
}
