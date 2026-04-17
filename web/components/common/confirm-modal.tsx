'use client';
import { Button, Dialog, Flex, Text } from '@radix-ui/themes';

export function ConfirmModal({
  open,
  title,
  description,
  confirmLabel = 'Confirm',
  onConfirm,
  onClose,
  destructive = false,
}: {
  open: boolean;
  title: string;
  description?: string;
  confirmLabel?: string;
  onConfirm: () => void | Promise<void>;
  onClose: () => void;
  destructive?: boolean;
}) {
  return (
    <Dialog.Root open={open} onOpenChange={(o) => !o && onClose()}>
      <Dialog.Content maxWidth="440px">
        <Dialog.Title>{title}</Dialog.Title>
        {description && (
          <Dialog.Description size="2" color="gray">
            {description}
          </Dialog.Description>
        )}
        <Flex gap="2" justify="end" mt="4">
          <Button variant="soft" color="gray" onClick={onClose}>
            Cancel
          </Button>
          <Button color={destructive ? 'red' : undefined} onClick={onConfirm}>
            {confirmLabel}
          </Button>
        </Flex>
      </Dialog.Content>
    </Dialog.Root>
  );
}
