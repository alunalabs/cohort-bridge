'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function AboutPage() {
    const router = useRouter();

    useEffect(() => {
        // Redirect to home page since about content is now there
        router.replace('/');
    }, [router]);

    return (
        <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-100 flex items-center justify-center">
            <div className="text-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
                <p className="text-slate-600">Redirecting to home page...</p>
            </div>
        </div>
    );
} 