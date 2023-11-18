<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageData, ActionData } from './$types';

	export let data: PageData;
	export let form: ActionData;
</script>

<section>
	<h2>Sign in</h2>
	<form use:enhance method="post" action="?/signin">
		<button>Sign in</button>
	</form>
</section>
