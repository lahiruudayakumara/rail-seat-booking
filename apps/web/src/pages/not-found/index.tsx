import { Helmet } from "react-helmet-pro";
import { Link } from "react-router-dom";
import { Button } from "@/components";

export function NotFoundPage() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-6 bg-[#f7f6f5] text-stone-900 text-center">
      <Helmet>
        <title>Page Not Found | Lanka Rail Reserve</title>
      </Helmet>
      <h1 className="font-heading text-6xl font-extrabold text-maroon-900">404</h1>
      <h2 className="mt-2 text-2xl font-bold text-stone-800">Page Not Found</h2>
      <p className="mt-2 text-stone-500 max-w-md">
        The page you are looking for does not exist or has been moved.
      </p>
      <Link to="/" className="mt-6">
        <Button variant="primary">Back to Home</Button>
      </Link>
    </div>
  );
}

export default NotFoundPage;
