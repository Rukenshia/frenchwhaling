<script>
  import { shipInfo } from './store';

  export let ship;

  function tierToName(tier) {
    switch (tier) {
      case 1:
        return 'I';
      case 2:
        return 'II';
      case 3:
        return 'III';
      case 4:
        return 'IV';
      case 5:
        return 'V';
      case 6:
        return 'VI';
      case 7:
        return 'VII';
      case 8:
        return 'VIII';
      case 9:
        return 'IX';
      case 10:
        return 'X';
      default:
        return '';
    }
  }
</script>

{#if $shipInfo}
  <div class="flex w-full">
    <div class="w-10">
      <img
        alt="image of {$shipInfo[ship.ship_id].name}"
        src={$shipInfo[ship.ship_id].images.small}
      />
    </div>
    <div
      class="w-auto pl-2 mt-1 pr-2 text-gray-500"
      class:line-through={ship.Resource.Earned}
      class:text-gray-400={ship.Resource.Earned}
    >
      <span class="invisible md:visible"
        >{tierToName($shipInfo[ship.ship_id].tier)}</span
      >
      {$shipInfo[ship.ship_id].name}
    </div>
  </div>
{/if}
