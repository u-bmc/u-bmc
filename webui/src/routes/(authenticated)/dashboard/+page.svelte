<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import { Separator } from "$lib/components/ui/separator/index.js";
    import * as Breadcrumb from "$lib/components/ui/breadcrumb/index.js";

	import { getContext, onDestroy } from "svelte";
	import type { GetDataResponse } from "../../../gen/proto/v1/data_pb";
	import { GetDataResponseSchema } from "../../../gen/proto/v1/data_pb";
	import { fromBinary } from "@bufbuild/protobuf";
	import type { NatsConnection } from "nats.ws";

	interface Response {
		message: string;
	}

	let data: Response = {
		message: ''
	};

	let lastRequestInfo = '';

	const nc: NatsConnection = getContext('natsClient');

	async function fetchData() {
		try {
			const sc = nc.services.client();
			let iter = await sc.stats();

			const req = await nc.request("datagen.gen", "");
			const responseData: GetDataResponse = fromBinary(GetDataResponseSchema, req.data);

			for await (const m of iter) {
				lastRequestInfo = m.id;
			}

			data = {
				message: responseData.message
			};
		} catch (err) {
			console.error(err);
			data = {
				message: ''
			};
		}
	}

	// Uncomment the following lines to fetch data periodically
	// $: {
	// 	fetchData();
	// 	const interval = setInterval(fetchData, 2000);
	// 	onDestroy(() => clearInterval(interval));
	// }
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
    <div class="bg-muted/50 min-h-[100vh] flex-1 rounded-xl md:min-h-min">
        <p>Data: {data.message}</p>
        <p>Last Request Info: {lastRequestInfo}</p>
    </div>
</div>


