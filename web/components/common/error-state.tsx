export function ErrorState({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return (
    <div className="p-8 text-center text-sm text-muted-foreground">
      <div className="mb-2">{message}</div>
      {onRetry && (
        <button
          onClick={onRetry}
          className="text-foreground underline underline-offset-4 hover:no-underline"
        >
          Try again
        </button>
      )}
    </div>
  );
}
