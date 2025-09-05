<script lang="ts">
	import AppSidebar from '$lib/components/app-sidebar.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';

	import { createConnectTransport } from '@connectrpc/connect-web';
	import { createClient } from '@connectrpc/connect';
	import { setContext } from 'svelte';
	import { BMCService } from '../../gen/schema/v1alpha1/system_pb';

	let { children } = $props();

	async function initializeConnectRPCClient() {
		try {
			const transport = createConnectTransport({
				baseUrl: `http://${window.location.hostname}:443`,
				useBinaryFormat: true
			});

			const client = createClient(BMCService, transport);

			setContext('rpc-client', client);
		} catch (err) {
			console.error('Failed to initialize ConnectRPC client:', err);
		}
	}

	initializeConnectRPCClient();
</script>

<Sidebar.Provider>
	<AppSidebar />
	<Sidebar.Inset>
		{@render children()}
	</Sidebar.Inset>
</Sidebar.Provider>
