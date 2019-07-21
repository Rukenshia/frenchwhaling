<script>
    import { derived, writable } from 'svelte/store';
    import { onMount } from 'svelte';
    import { dataUrl, shipInfo } from './store';
    import moment from 'moment';
    import axios from 'axios';
    import ShipInfo from './ShipInfo.svelte';

    export let isNew;
    let retries = 0;

    let data = writable(undefined);
    let error = false;
    const max = derived(data, v => {
        if (v === undefined) { return [0, 0]; }

        const newMax = [0, 0];
        Object.keys(v.Ships).forEach(s => {
            newMax[v.Ships[s].Resource.Type] += v.Ships[s].Resource.Amount;
        });

        return newMax;
    }, [0, 0]);
    const categories = derived([data, shipInfo], ([v, vs]) => {
        if (v === undefined || vs === undefined) { return [{}, {}]; }

        const sort = (a, b) => {
            const byName = () => {
                if (vs[a.ship_id].name < vs[b.ship_id].name) {
                    return -1;
                } else  if (vs[a.ship_id].name > vs[b.ship_id].name) {
                    return 1;
                }
                return 0;
            }
            if (a.Resource.Earned) {
                if (b.Resource.Earned) {
                    return byName();
                }
                return 1;
            } else {
                if (!b.Resource.Earned) {
                    return byName();
                }
                return -1;
            }
        };

        return [
            Object.values(v.Ships).filter(s => s.Resource.Type === 0).sort(sort).reduce((agg, x) => {
                if (!agg[x.Resource.Amount]) {
                    agg[x.Resource.Amount] = { Amount: x.Resource.Amount, Ships: [] };
                }

                agg[x.Resource.Amount].Ships.push(x);
                return agg;
            }, {}),
            Object.values(v.Ships).filter(s => s.Resource.Type === 1).sort(sort).reduce((agg, x) => {
                if (!agg[x.Resource.Amount]) {
                    agg[x.Resource.Amount] = { Amount: x.Resource.Amount, Ships: [] };
                }

                agg[x.Resource.Amount].Ships.push(x);
                return agg;
            }, {}),
        ];
    }, [{}, {}])

    const resourceName = ['Republic Tokens', 'Coal'];

    onMount(async () => {
        try {
            const res = await axios.get($dataUrl);
            $data = res.data;
        } catch(e) {
            error = true;

            if (isNew) {
                const intv = setInterval(async () => {
                    retries++;
                    console.log(retries);

                    if (retries > 9) {
                        clearInterval(intv);
                        return;
                    }

                    try {
                        const res = await axios.get($dataUrl);
                        $data = res.data;
                        error = false;
                    } catch(e) {
                        error = true;
                    }
                }, 2500);
            }
        }
    });
</script>

{#if $data}

<div class="ml-4 text-gray-600 font-medium text-sm">
    Last updated {moment($data.LastUpdated / 1000000).fromNow()}
</div>
<div class="w-full flex flex-wrap mt-4 px-2">
{#each $data.Resources as resource}
    <div class="w-full md:w-1/2">
        <div class="m-2 p-4 shadow-xl rounded bg-gray-200">
            <div class="flex">
                <div class="w-7">
                    <img class="w-8" alt="resource icon" src="/img/resources/{resource.Type}.png" />
                </div>
                <div class="w-auto ml-2 text-xl text-gray-700">{resourceName[resource.Type]}</div>
            </div>
            <div class="p-4 text-gray-700">
                You have earned up to <span class="text-3xl">{resource.Earned}</span> {resourceName[resource.Type]} out of <span class="text-3xl">{$max[resource.Type]}</span> you can earn during the event.
            </div>


        {#if $categories}
        {#each Object.keys($categories[resource.Type]).reverse() as amount}
            <div class="flex flex-wrap mb-4">
                <div class="w-full pl-2 text-sm text-gray-600 font-medium">{amount} {resourceName[resource.Type]}</div>
                {#each $categories[resource.Type][amount].Ships as ship}
                <div class="w-1/2 xl:w-2/6 p-1 ">
                    <div class="border-2 rounded" class:border-green-200={ship.Resource.Earned > 0}>
                        <ShipInfo {ship} />
                    </div>
                </div>
                {/each}
            </div>
        {/each}
        {/if}
        </div>

    </div>
{/each}
</div>

{:else}
<div class="w-full flex flex-wrap justify-around mt-4">
{#if isNew && retries <= 9}
    <div class="w-full text-center text-5xl text-gray-600 mt-8">
        Loading your stuff
    </div>
    <div class="w-1/2 text-center text-2xl text-gray-500">
        You're apparently new here. That's cool. Loading your data might take a bit depending on the server load.
        Just stay put.
    </div>
    <div class="w-3/4 text-center text-xs text-gray-500 font-mono">
        attempt {retries + 1} of 10
    </div>
{/if}
{#if error && retries > 9}
    <div class="w-full text-center text-6xl text-red-700 rounded font-mono mt-8">Big Red Error</div>
    <div class="w-3/4 text-center text-2xl text-gray-800 rounded font-mono">
        There are a lot of things that can go wrong. Guess what, you're a lucky one. You've caught the
        big red error. Basically, nothing works.
        <br />
        A refresh of the page might help.
    </div>
{/if}
    </div>
{/if}

