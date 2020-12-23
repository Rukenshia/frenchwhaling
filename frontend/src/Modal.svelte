<script>
  import { createEventDispatcher } from "svelte";
  import { fade, scale } from "svelte/transition";
  import { cubicIn, cubicOut } from "svelte/easing";
  export let show;
  export let title;
  export let message;

  const dispatch = createEventDispatcher();
</script>

{#if show}
  <div class="fixed z-10 inset-0 overflow-y-auto">
    <div
      class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
      <div
        class="fixed inset-0 transition-opacity"
        aria-hidden="true"
        in:fade={{ duration: 300, start: 0.0, easing: cubicOut }}
        out:fade={{ duration: 200, start: 1, easing: cubicIn }}>
        <div class="absolute inset-0 bg-gray-500 opacity-75" />
      </div>

      <!-- This element is to trick the browser into centering the modal contents. -->
      <span
        class="hidden sm:inline-block sm:align-middle sm:h-screen"
        aria-hidden="true">&#8203;</span>
      <!--
        Modal panel, show/hide based on modal state.
  
        Entering: "ease-out duration-300"
          From: "opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
          To: "opacity-100 translate-y-0 sm:scale-100"
        Leaving: "ease-in duration-200"
          From: "opacity-100 translate-y-0 sm:scale-100"
          To: "opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
      -->
      <div
        class="inline-block align-bottom bg-gray-800 rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-sm sm:w-full sm:p-6"
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-headline"
        in:scale={{ duration: 300, start: 0.95, easing: cubicOut }}
        out:scale={{ duration: 200, start: 0.95, easing: cubicIn }}>
        <div>
          <div class="text-center">
            <h3
              class="text-lg leading-6 font-medium text-gray-200"
              id="modal-headline">
              {title}
            </h3>
            <div class="mt-2">
              <p class="text-sm text-gray-300">
                {@html message}
              </p>
            </div>
          </div>
        </div>
        <div class="mt-5 sm:mt-6">
          <button
            type="button"
            on:click={() => dispatch('close')}
            class="inline-flex justify-center w-full rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:text-sm">
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
