<script lang="ts">
	let name = 'world';
	import { writable } from 'svelte/store';

	const messages = writable<Array<{ body: string; id: number }>>([]);
	const evtSource = new EventSource('http://localhost:4000/v1/events');
	evtSource.onerror = function (err) {
		console.log(err);
	};

	evtSource.onopen = function (event) {
		console.log(event, 'opened!!');
	};

	evtSource.addEventListener('post-added', function (event) {
		const data: { body: string; id: number } = JSON.parse(event.data);
		console.log('data: ', data);
		messages.set($messages.concat(data));
	});
</script>

<h1>Hello {name}!</h1>

<details>
	<summary>How this works</summary>
	<small
		>This site opens a Server-Sent Event (SSE) Source and then updates the page when the server
		sends any events. Default event time on sse.dev is 2 seconds<br />
		See more at <a target="_blank" href="https://sse.dev/">https://sse.dev/</a></small
	>
</details>

{#each $messages as m (m.id)}
	<p>
		{m.body}
	</p>
{/each}
