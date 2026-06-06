<script>
  export let show = false
  export let title = ''
  export let message = ''
  export let type = 'alert' // 'alert' | 'confirm'
  export let confirmText = '确定'
  export let cancelText = '取消'

  export let resolve

  function handleConfirm() {
    show = false
    if (resolve) resolve(true)
  }

  function handleCancel() {
    show = false
    if (resolve) resolve(false)
  }

  function handleOverlay() {
    if (type !== 'confirm') {
      show = false
      if (resolve) resolve(false)
    }
  }

  function handleKeydown(e) {
    if (!show) return
    if (e.key === 'Escape') {
      if (type === 'confirm') {
        handleCancel()
      } else {
        handleConfirm()
      }
    }
    if (e.key === 'Enter') {
      handleConfirm()
    }
  }

  $: if (show) {
    document.addEventListener('keydown', handleKeydown)
  } else {
    document.removeEventListener('keydown', handleKeydown)
  }
</script>

{#if show}
  <div class="modal-overlay" on:click={handleOverlay} role="dialog" aria-modal="true">
    <div class="modal" on:click|stopPropagation role="document">
      <div class="modal-header">
        <h3>{title}</h3>
      </div>
      <div class="modal-body">
        <p class="message">{message}</p>
      </div>
      <div class="modal-footer">
        {#if type === 'confirm'}
          <button class="btn btn-secondary" on:click={handleCancel}>{cancelText}</button>
        {/if}
        <button class="btn btn-primary" on:click={handleConfirm}>{confirmText}</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2000;
    animation: fadeIn 0.2s ease;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .modal {
    background: white;
    border-radius: 16px;
    width: 400px;
    max-width: 90vw;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.25);
    animation: slideUp 0.25s ease;
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .modal-header {
    padding: 20px 24px 12px;
  }

  .modal-header h3 {
    font-size: 1.15rem;
    color: #2c1810;
    margin: 0;
  }

  .modal-body {
    padding: 8px 24px 20px;
  }

  .message {
    font-size: 0.95rem;
    color: #5a4636;
    line-height: 1.6;
    white-space: pre-line;
  }

  .modal-footer {
    padding: 16px 24px;
    border-top: 1px solid #f0e8e0;
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }

  .btn {
    padding: 8px 20px;
    border: none;
    border-radius: 8px;
    font-size: 0.9rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-primary {
    background: #8b2500;
    color: white;
  }
  .btn-primary:hover { background: #a0522d; }

  .btn-secondary {
    background: #f5f0eb;
    color: #5a4636;
    border: 1px solid #d4c4b5;
  }
  .btn-secondary:hover { background: #ebe4dc; }
</style>
