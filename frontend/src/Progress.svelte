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
    let withShipsNotInGarage = false;
    let resource = writable();

    const timestamp = writable(+ new Date());
    const lastUpdatedMoment = writable(undefined);

    let data = writable(undefined);
    let initialDataLoaded = writable(false);
    let error = false;
    const max = derived(data, v => {
        if (v === undefined) { return [0, 0]; }

        const newMax = [[0, 0], [0, 0], [0, 0], [0, 0]];
        Object.keys(v.Ships).forEach(s => {
            if (v.Ships[s].private.in_garage) {
                newMax[v.Ships[s].Resource.Type][0] += v.Ships[s].Resource.Amount;
            }
            newMax[v.Ships[s].Resource.Type][1] += v.Ships[s].Resource.Amount;
        });

        return newMax;
    }, [[0, 0], [0, 0], [0, 0], [0, 0]]);
    const categories = derived([data, shipInfo], ([v, vs]) => {
        if (v === undefined || vs === undefined) { return [{}, {}, {}, {}]; }

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

        const getCategory = resourceType => {
            return Object.values(v.Ships).filter(s => s.Resource.Type === resourceType).sort(sort).reduce((agg, x) => {
                if (!agg[x.Resource.Amount]) {
                    agg[x.Resource.Amount] = { Amount: x.Resource.Amount, Ships: [] };
                }

                agg[x.Resource.Amount].Ships.push(x);
                return agg;
            }, {});
        };

        return [
            getCategory(0),
            getCategory(1),
            getCategory(2),
            getCategory(3),
        ];
    }, [{}, {}, {}, {}]);

    const resourceName = ['Republic Tokens', 'Coal', 'Steel', 'Santa Container'];

    function refresh() {
        axios.get(`https://whaling-api.in.fkn.space/subscribers/${$accountId}/refresh`, {
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

    async function reloadDataWithRetry(tries = 60, done = () => {}) {
        let lastUpdated = undefined;
        if ($data) {
            reloading = true;
            lastUpdated = $data.LastUpdated;
        }
        retries = 0;

        try {
            const res = await axios.get(`${$dataUrl}?${+ new Date()}`);
            $data = res.data;
            done();

            if (reloading) {
                if (lastUpdated >= $data.LastUpdated) {
                    throw new Error('Not updated yet');
                }
            }
            reloading = false;
        } catch(e) {
            error = true;
            console.log(e);

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
                    done();

                    if (reloading) {
                        if (lastUpdated >= $data.LastUpdated) {
                            throw new Error('Not updated yet');
                        }
                    }
                    
                    error = false;
                    reloading = false;
                    clearInterval(intv);
                } catch(e) {
                    console.log(e);
                    error = true;
                }
            }, 2500);
        }
    }

    onMount(async () => {
        $timestamp = +new Date() * 1000000;

        await reloadDataWithRetry(60, () => {
            $resource = $data.Resources[1];
            $lastUpdatedMoment = moment($data.LastUpdated / 1000000).fromNow();

            setInterval(() => {
                $timestamp = +new Date() * 1000000;
                $lastUpdatedMoment = moment($data.LastUpdated / 1000000).fromNow();
            }, 2500);
        });
    });
</script>

<style>
button.cursor-not-allowed {
    @apply bg-gray-900;
}
</style>

{#if $data}

<div class="ml-4 text-gray-400 font-medium text-sm">
    {#if reloading}
    <span class="font-mono">Loading...</span>
    {:else}
    Last updated {$lastUpdatedMoment}

    {#if $timestamp - $data.LastUpdated > 10 * 60 * 1000 * 1000000}
        <button on:click={refresh} class="text-gray-200 bg-gray-700 hover:bg-gray-800 border-none rounded shadow-md p-2">Refresh now</button>
    {:else}
        <button disabled class="border-2 border-gray-600 rounded p-2 cursor-not-allowed">Refresh now</button>
    {/if}
    {/if}
</div>
<div class="w-full flex flex-wrap mt-4 px-2">
{#each $data.Resources.slice(1) as res}
    <div class="w-1/3" on:click={() => $resource = res}>
        <div style="transition: background-color .1s" class:bg-blue-900={res === $resource} class="m-2 shadow-xl rounded rounded-b-none bg-gray-800 overflow-hidden hover:bg-blue-900 hover:shadow-md">
            <div class="p-4 pb-2 flex">
                <div class="w-7">
                    <img class="h-8 w-auto" alt="resource" src="/img/resources/{res.Type}.png" />
                </div>
                <div class="sm:hidden w-auto ml-2 text-lg text-gray-400">
                    {res.Earned / Math.max(1, $max[res.Type][withShipsNotInGarage ? 1 : 0]) * 100}%
                </div>
                <div class="hidden sm:block w-auto ml-2 text-lg text-gray-400">{res.Earned} of {$max[res.Type][withShipsNotInGarage ? 1 : 0]}</div>
            </div>
            <div class="relative h-2 w-full z-0 bg-gray-700">
                <div style="width: {res.Earned / Math.max(1, $max[res.Type][withShipsNotInGarage ? 1 : 0]) * 100}%" class="absolute bottom-0 h-2 bg-green-900"></div>
            </div>
        </div>
    </div>
{/each}
</div>
<div class="w-full flex flex-wrap -mt-4 px-2">
    <div class="w-full">
        <div class="m-2 p-4 shadow-xl rounded-t-none rounded bg-gray-800">
        {#if $resource}
            <div class="flex">
                <div class="w-7">
                    <img class="h-8 w-auto" alt="resource" src="/img/resources/{$resource.Type}.png" />
                </div>
                <div class="w-auto ml-2 text-xl text-gray-400">{resourceName[$resource.Type]}</div>
            </div>
            <div class="p-4 text-gray-500">
                You have earned up to <span class="text-3xl">{$resource.Earned}</span> {resourceName[$resource.Type]} out of <span class="text-3xl">{$max[$resource.Type][withShipsNotInGarage ? 1 : 0]}</span> you can earn during the event.
            </div>
            <div class="p-4 pt-0">
                <label class="md:w-full block text-gray-600 font-bold">
                    <input class="mr-2 leading-tight" type="checkbox" bind:checked={withShipsNotInGarage}>
                    <span class="text-sm">
                        Include ships I used to have in port
                    </span>
                </label>
            </div>

            {#if $categories}
            {#each Object.keys($categories[$resource.Type]).reverse() as amount}
                <div class="flex flex-wrap mb-4">
                    <div class="w-full pl-2 text-sm text-gray-600 font-medium">{amount} {resourceName[$resource.Type]}</div>
                    {#each $categories[$resource.Type][amount].Ships as ship}
                    {#if withShipsNotInGarage || ship.private.in_garage}
                    <div class="w-1/2 lg:w-1/2 xl:w-1/4 p-1 ">
                        <div class="border-2 border-gray-600 rounded" class:border-green-900={ship.Resource.Earned > 0} class:border-yellow-800={ship.private && !ship.private.in_garage}>
                            <ShipInfo {ship} />
                        </div>
                    </div>
                    {/if}
                    {/each}
                </div>
            {/each}
            {/if}
        {/if}
        </div>

    </div>
</div>

{:else}
<div class="w-full flex flex-wrap justify-around mt-4">
{#if retries <= 9}
    <div class="w-full text-center text-5xl text-gray-400 mt-8">
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

