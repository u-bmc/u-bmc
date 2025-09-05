<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import { ModeWatcher } from 'mode-watcher';

	import { WebTracerProvider } from '@opentelemetry/sdk-trace-web';
	import { getWebAutoInstrumentations } from '@opentelemetry/auto-instrumentations-web';
	import { DocumentLoadInstrumentation } from '@opentelemetry/instrumentation-document-load';
	import { ZoneContextManager } from '@opentelemetry/context-zone';
	import { registerInstrumentations } from '@opentelemetry/instrumentation';
	import { B3Propagator } from '@opentelemetry/propagator-b3';

	// Create a tracer provider without any span exporters since spans will be
	// attached to Connect RPC API call headers instead of being logged to console
	const provider = new WebTracerProvider();

	provider.register({
		contextManager: new ZoneContextManager(),
		// Using B3 propagator to ensure trace context is properly passed in headers
		propagator: new B3Propagator()
	});

	registerInstrumentations({
		instrumentations: [new DocumentLoadInstrumentation(), ...getWebAutoInstrumentations()]
	});

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<ModeWatcher />

{@render children?.()}
