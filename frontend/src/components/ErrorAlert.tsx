interface ErrorAlertProps {
  message: string;
  onDismiss: () => void;
  className?: string;
}

export const ErrorAlert = ({ message, onDismiss, className = '' }: ErrorAlertProps) => (
  <div className={`rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 flex justify-between items-start ${className}`}>
    <span>{message}</span>
    <button onClick={onDismiss} className="ml-3 text-red-400 hover:text-red-600 flex-shrink-0">&times;</button>
  </div>
);
