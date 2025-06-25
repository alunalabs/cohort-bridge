import { ReactNode } from 'react';

interface ConfigLayoutProps {
    children: ReactNode;
}

export default function ConfigLayout({ children }: ConfigLayoutProps) {
    return (
        <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
            {children}
        </div>
    );
} 