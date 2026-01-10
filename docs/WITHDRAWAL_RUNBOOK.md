# Withdrawal Runbook

This runbook describes the recommended steps for safely executing and validating withdrawals.

## Before initiating a withdrawal

- Confirm the target network and recipient address.
- Verify the withdrawal amount and token decimals.
- Ensure the withdrawer service is fully synced.
- Check recent logs for errors or stalled processes.

## Initiating the withdrawal

- Submit the withdrawal transaction using the configured signer.
- Record the transaction hash immediately after submission.
- Avoid submitting duplicate requests for the same withdrawal.

## Monitoring progress

- Track the transaction status on the relevant block explorer.
- Monitor withdrawer logs for confirmation or retry messages.
- Ensure no unexpected reverts or retries occur.

## Post-withdrawal checks

- Verify the recipient balance increased by the expected amount.
- Confirm internal state (queues, pending withdrawals) is updated.
- Store the transaction hash and timestamp for audit purposes.

## Common issues and mitigations

- Insufficient gas: re-submit with updated gas parameters.
- Nonce conflicts: pause new withdrawals until resolved.
- RPC instability: retry only after confirming the previous attempt failed.

This runbook is intended to reduce operational risk and provide a clear checklist
for both routine and exceptional withdrawal operations.
