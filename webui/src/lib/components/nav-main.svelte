<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';

	let {
		items
	}: {
		items: {
			title: string;
			url: string;
			description: string;
			group: string;
			// this should be `Component` after lucide-svelte updates types
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			icon: any;
		}[];
	} = $props();
</script>

{#each ['Monitoring', 'Hardware Control', 'Settings'] as groupLabel}
	<Sidebar.Group>
		<Sidebar.GroupLabel>{groupLabel}</Sidebar.GroupLabel>
		<Sidebar.Menu>
			{#each items as item (item.title)}
				{#if item.group === groupLabel.toLowerCase()}
					<Sidebar.MenuItem>
						<Sidebar.MenuButton>
							{#snippet child({ props })}
								<a href={item.url} {...props}>
									{#snippet tooltipContent()}
										{item.description}
									{/snippet}
									{#if item.icon}
										<item.icon />
									{/if}
									<span>{item.title}</span>
								</a>
							{/snippet}
						</Sidebar.MenuButton>
					</Sidebar.MenuItem>
				{/if}
			{/each}
		</Sidebar.Menu>
	</Sidebar.Group>
{/each}
