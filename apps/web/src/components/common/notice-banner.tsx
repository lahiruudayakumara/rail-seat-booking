interface NoticeBannerProps {
  message: string;
}

export function NoticeBanner({ message }: NoticeBannerProps) {
  if (!message) return null;
  return (
    <div className="notice mt-5" role="alert">
      {message}
    </div>
  );
}

interface InlineErrorProps {
  message: string;
}

export function InlineError({ message }: InlineErrorProps) {
  return (
    <p className="mt-3 text-sm text-red-700" role="alert">
      {message}
    </p>
  );
}
