<script>
    import axios from 'axios';
    import { shipInfo } from './store';

    const applicationId = 'bf7fb7e1809acb24157245221cca089b';

    export let ship;
</script>

{#if $shipInfo}
<div class="flex w-full">
    <div class="w-10 pl-1">
        <img alt="image of {$shipInfo[ship.ship_id].name}"src={$shipInfo[ship.ship_id].images.small} />
    </div>
    <div class="w-auto pl-2 mt-1" class:text-gray-600={ship.Resource.Earned}>
        {$shipInfo[ship.ship_id].name}
    </div>
</div>
{/if}