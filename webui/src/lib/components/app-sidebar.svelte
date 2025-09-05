<script lang="ts" module>
	import Gauge from 'lucide-svelte/icons/gauge';
	import HardDriveDownload from 'lucide-svelte/icons/hard-drive-download';
	import Logs from 'lucide-svelte/icons/logs';
	import MonitorUp from 'lucide-svelte/icons/monitor-up';
	import Network from 'lucide-svelte/icons/network';
	import Power from 'lucide-svelte/icons/power';
	import Server from 'lucide-svelte/icons/server';
	import ServerCog from 'lucide-svelte/icons/server-cog';
	import ServerCrash from 'lucide-svelte/icons/server-crash';
	import Settings from 'lucide-svelte/icons/settings';
	import SquareTerminal from 'lucide-svelte/icons/square-terminal';
	import ThermometerSnowflake from 'lucide-svelte/icons/thermometer-snowflake';
	import USB from 'lucide-svelte/icons/usb';
	import UserCog from 'lucide-svelte/icons/user-cog';

	const data = {
		user: {
			name: 'admin',
			role: 'Administrator'
		},
		bmcs: [
			{
				name: 'Dummy BMC 0',
				icon: Server,
				platform: 'Tyan S8030'
			},
			{
				name: 'Dummy BMC 1',
				icon: Server,
				platform: 'Tyan S8040'
			},
			{
				name: 'Dummy BMC 2',
				icon: Server,
				platform: 'Tyan S5549'
			}
		],
		navMain: [
			{
				title: 'Dashboard',
				url: '/dashboard',
				description: 'BMC and server status',
				group: 'monitoring',
				icon: Gauge
			},
			{
				title: 'Logs',
				url: '/logs',
				description: 'Server event and post logs',
				group: 'monitoring',
				icon: Logs
			},
			{
				title: 'KVM',
				url: '/kvm',
				description: 'KVM remote control',
				group: 'hardware control',
				icon: MonitorUp
			},
			{
				title: 'SoL',
				url: '/sol',
				description: 'Serial over LAN',
				group: 'hardware control',
				icon: SquareTerminal
			},
			{
				title: 'Power Control',
				url: '/power-control',
				description: 'BMC and Host power control',
				group: 'hardware control',
				icon: Power
			},
			{
				title: 'Virtual Media',
				url: '/virtual-media',
				description: 'Virtual media management',
				group: 'hardware control',
				icon: USB
			},
			{
				title: 'Firmware Update',
				url: '/firmware-update',
				description: 'System firmware management',
				group: 'hardware control',
				icon: HardDriveDownload
			},
			{
				title: 'Misceallaneous',
				url: '/misc',
				description: 'Miscellaneous hardware control',
				group: 'hardware control',
				icon: ServerCog
			},
			{
				title: 'Power',
				url: '/power-settings',
				description: 'Power limit and restore settings',
				group: 'settings',
				icon: ServerCrash
			},
			{
				title: 'Thermal',
				url: '/thermal-settings',
				description: 'Thermal settings',
				group: 'settings',
				icon: ThermometerSnowflake
			},
			{
				title: 'Network',
				url: '/network-settings',
				description: 'Network settings',
				group: 'settings',
				icon: Network
			},
			{
				title: 'User Management',
				url: '/user-management',
				description: 'User management',
				group: 'settings',
				icon: UserCog
			},
			{
				title: 'System Settings',
				url: '/settings',
				description: 'System settings',
				group: 'settings',
				icon: Settings
			}
		]
	};
</script>

<script lang="ts">
	import NavMain from '$lib/components/nav-main.svelte';
	import NavUser from '$lib/components/nav-user.svelte';
	import TeamSwitcher from '$lib/components/bmc-switcher.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import type { ComponentProps } from 'svelte';

	let {
		ref = $bindable(null),
		collapsible = 'icon',
		...restProps
	}: ComponentProps<typeof Sidebar.Root> = $props();
</script>

<Sidebar.Root bind:ref {collapsible} {...restProps}>
	<Sidebar.Header>
		<TeamSwitcher bmcs={data.bmcs} />
	</Sidebar.Header>
	<Sidebar.Content>
		<NavMain items={data.navMain} />
	</Sidebar.Content>
	<Sidebar.Footer>
		<NavUser user={data.user} />
	</Sidebar.Footer>
	<Sidebar.Rail />
</Sidebar.Root>
