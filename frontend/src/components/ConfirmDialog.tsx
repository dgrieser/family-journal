import { useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Trash2 } from 'lucide-react';

interface ConfirmDialogProps {
  open: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

export const ConfirmDialog = ({ open, onConfirm, onCancel }: ConfirmDialogProps) => {
  const { t } = useTranslation();

  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onCancel();
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [open, onCancel]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4" role="dialog" aria-modal="true">
      <div
        className="absolute inset-0 bg-black/30 backdrop-blur-sm"
        onClick={onCancel}
      />
      <div className="confirm-dialog-panel relative bg-white rounded-xl shadow-2xl border border-stone-200 p-6 w-full max-w-sm">
        <div className="flex gap-4 mb-6">
          <div className="flex-shrink-0 w-10 h-10 rounded-full bg-red-50 border border-red-100 flex items-center justify-center">
            <Trash2 size={16} className="text-red-500" />
          </div>
          <div className="pt-0.5">
            <h3 className="text-base font-semibold text-stone-900">{t('confirm_delete')}</h3>
            <p className="text-sm text-stone-500 mt-1">{t('confirm_delete_body')}</p>
          </div>
        </div>
        <div className="flex justify-end gap-2.5">
          <button
            onClick={onCancel}
            className="px-4 py-2 text-sm rounded-lg border border-stone-200 text-stone-600 hover:bg-stone-50 transition-colors font-medium"
          >
            {t('cancel')}
          </button>
          <button
            onClick={onConfirm}
            autoFocus
            className="px-4 py-2 text-sm rounded-lg font-medium text-white bg-red-600 hover:bg-red-500 transition-colors inline-flex items-center gap-1.5"
          >
            <Trash2 size={14} />
            {t('delete')}
          </button>
        </div>
      </div>
    </div>
  );
};
