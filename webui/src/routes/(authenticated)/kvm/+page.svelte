<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb/index.js';

	import RFB from '@novnc/novnc/lib/rfb';
	import type { Action } from 'svelte/action';

	const KvmAction: Action = (node) => {
		$effect(() => {
			let rfb: RFB = new RFB(node, `ws://${window.location.hostname}:5900`, {
				scaleViewport: true,
				resizeSession: true
			});
			return () => rfb.disconnect();
		});
	};
</script>

<header
	class="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12"
>
	<div class="flex items-center gap-2 px-4">
		<Sidebar.Trigger class="-ml-1" />
		<Separator orientation="vertical" class="mr-2 h-4" />
		<Breadcrumb.Root>
			<Breadcrumb.List>
				<Breadcrumb.Item class="hidden md:block">
					<Breadcrumb.Link href="#">Monitoring</Breadcrumb.Link>
				</Breadcrumb.Item>
				<Breadcrumb.Separator class="hidden md:block" />
				<Breadcrumb.Item>
					<Breadcrumb.Page>Dashboard</Breadcrumb.Page>
				</Breadcrumb.Item>
			</Breadcrumb.List>
		</Breadcrumb.Root>
	</div>
</header>
<div class="flex flex-1 flex-col gap-4 p-4 pt-0">
	<div class="min-h-[100vh] flex-1 rounded-xl bg-muted/50 md:min-h-min">
		<p>Data: {data.message}</p>
		<p>Last Request Info: {lastRequestInfo}</p>
	</div>
</div>
