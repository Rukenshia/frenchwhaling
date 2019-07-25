<script>
    import { derived, writable } from 'svelte/store';
    import { onMount } from 'svelte';
    import { accountId, dataUrl, token, shipInfo } from './store';
    import moment from 'moment';
    import axios from 'axios';
    import ShipInfo from './ShipInfo.svelte';

    export let isNew;
    let retries = 0;
    let reloading = false;
    let withShipsNotInGarage = [false, false];

    const timestamp = writable(+ new Date());
    const lastUpdatedMoment = writable(undefined);

    let data = writable(undefined);
    let error = false;
    const max = derived(data, v => {
        if (v === undefined) { return [0, 0]; }

        const newMax = [[0, 0], [0, 0]];
        Object.keys(v.Ships).forEach(s => {
            if (v.Ships[s].private.in_garage) {
                newMax[v.Ships[s].Resource.Type][0] += v.Ships[s].Resource.Amount;
            }
            newMax[v.Ships[s].Resource.Type][1] += v.Ships[s].Resource.Amount;
        });

        return newMax;
    }, [[0, 0], [0, 0]]);
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
            };
            const byInGarage = () => {
                if (a.private.in_garage) {
                    if (b.private.in_garage) {
                        return byName();
                    }

                    return -1;
                } else {
                    if (b.private.in_garage) {
                        return 1;
                    }
                    return byName();
                }
            }
            if (a.Resource.Earned) {
                if (b.Resource.Earned) {
                    return byName();
                }
                return 1;
            } else {
                if (!b.Resource.Earned) {
                    return byInGarage();
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
    }, [{}, {}]);

    const resourceName = ['Republic Tokens', 'Coal'];

    function refresh() {
        axios.get(`https://frenchwhaling-api.in.fkn.space/subscribers/${$accountId}/refresh`, {
            headers: {
                'Authorization': `Bearer ${$token}`,
            },
        })
            .then(res => {
                reloadDataWithRetry(60);
            })
            .catch(err => {
                console.log(err, err.response);
                alert('Sorry, we could not refresh your data at this time. Please try logging out and in again, if that still does not work please contact me. Your data is also updated automatically every hour');
            });
    }

    async function reloadDataWithRetry(tries = 60) {
        let lastUpdated = undefined;
        if ($data) {
            reloading = true;
            lastUpdated = $data.LastUpdated;
        }
        retries = 0;

        try {
            const res = await axios.get(`${$dataUrl}?${+ new Date()}`);
            $data = res.data;

            if (reloading) {
                if (lastUpdated >= $data.LastUpdated) {
                    throw new Error('Not updated yet');
                }
            }
            reloading = false;
        } catch(e) {
            error = true;

            const intv = setInterval(async () => {
                console.log('retry');
                retries++;

                if (retries > tries) {
                    clearInterval(intv);
                    reloading = false;
                    return;
                }

                try {
                    const res = await axios.get(`${$dataUrl}?${+ new Date()}`);
                    $data = res.data;

                    if (reloading) {
                        if (lastUpdated >= $data.LastUpdated) {
                            throw new Error('Not updated yet');
                        }
                    }
                    
                    error = false;
                    reloading = false;
                    clearInterval(intv);
                } catch(e) {
                    error = true;
                }
            }, 2500);
        }
    }

    onMount(async () => {
        $timestamp = +new Date() * 1000000;
        await reloadDataWithRetry();
        $lastUpdatedMoment = moment($data.LastUpdated / 1000000).fromNow();

        setInterval(() => {
            $timestamp = +new Date() * 1000000;
            $lastUpdatedMoment = moment($data.LastUpdated / 1000000).fromNow();
        }, 2500);
    });
</script>

{#if $data}

<div class="ml-4 text-gray-600 font-medium text-sm">
    {#if reloading}
    <span class="font-mono">Loading...</span>
    {:else}
    Last updated {$lastUpdatedMoment}

    {#if $timestamp - $data.LastUpdated > 10 * 60 * 1000 * 1000000}
        <button on:click={refresh} class="bg-gray-200 hover:bg-gray-300 border-none rounded shadow-md p-2">Refresh now</button>
    {:else}
        <button disabled class="border-2 border-gray-200 rounded p-2 cursor-not-allowed">Refresh now</button>
    {/if}
    {/if}
</div>
<div class="w-full flex flex-wrap mt-4 px-2">
{#each $data.Resources as resource}
    <div class="w-full lg:w-1/2">
        <div class="m-2 p-4 shadow-xl rounded bg-gray-200">
            <div class="flex">
                <div class="w-7">
                    <img class="w-8" alt="resource icon" src="/img/resources/{resource.Type}.png" />
                </div>
                <div class="w-auto ml-2 text-xl text-gray-700">{resourceName[resource.Type]}</div>
            </div>
            <div class="p-4 text-gray-700">
                You have earned up to <span class="text-3xl">{resource.Earned}</span> {resourceName[resource.Type]} out of <span class="text-3xl">{$max[resource.Type][withShipsNotInGarage[resource.Type] ? 1 : 0]}</span> you can earn during the event.
            </div>
            <div class="p-4 pt-0">
                <label class="md:w-2/3 block text-gray-600 font-bold">
                    <input class="mr-2 leading-tight" type="checkbox" bind:checked={withShipsNotInGarage[resource.Type]}>
                    <span class="text-sm">
                        Include ships I don't have in port
                    </span>
                </label>
            </div>


        {#if $categories}
        {#each Object.keys($categories[resource.Type]).reverse() as amount}
            <div class="flex flex-wrap mb-4">
                <div class="w-full pl-2 text-sm text-gray-600 font-medium">{amount} {resourceName[resource.Type]}</div>
                {#each $categories[resource.Type][amount].Ships as ship}
                {#if withShipsNotInGarage[resource.Type] || ship.private.in_garage}
                <div class="w-1/2 lg:w-1/2 xl:w-1/3 p-1 ">
                    <div class="border-2 rounded" class:border-green-200={ship.Resource.Earned > 0} class:border-red-200={ship.private && !ship.private.in_garage}>
                        <ShipInfo {ship} />
                    </div>
                </div>
                {/if}
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
{#if retries <= 9}
    <div class="w-full text-center text-5xl text-gray-600 mt-8">
        Loading your stuff
    </div>
    <div class="w-1/2 text-center text-2xl text-gray-500">
        {#if isNew}You're apparently new here. That's cool.{:else}Welcome back, fellow whale.{/if} Loading your data might take a bit depending on the server load.
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
        A refresh of the page might help. Otherwise, dunno... There's a contact button above and you can also reach me on various discord servers as Rukenshia#4396. I'm happy to try and help out.
    </div>
{/if}
    </div>
    <div class="mb-64"></div>
{/if}

