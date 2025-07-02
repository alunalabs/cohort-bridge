'use client';

import { useFormContext } from 'react-hook-form';
import { Shield } from 'lucide-react';

interface SecuritySectionProps {
    missingFields?: string[];
}

export default function SecuritySection({ missingFields = [] }: SecuritySectionProps) {
    const { register } = useFormContext();

    const getInputClass = (fieldName: string) => {
        const baseClass = "w-full px-3 py-2 border rounded-lg focus:ring-2 text-slate-900 placeholder-slate-400 bg-white transition-colors";
        const isMissing = missingFields.includes(fieldName);

        if (isMissing) {
            return `${baseClass} border-red-300 bg-red-50 focus:ring-red-500 focus:border-red-500`;
        }

        return `${baseClass} border-slate-300 focus:ring-blue-500 focus:border-blue-500`;
    };

    return (
        <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-6">
            <div className="mb-6">
                <div className="flex items-center space-x-3 mb-2">
                    <div className="w-8 h-8 bg-red-100 rounded-lg flex items-center justify-center">
                        <Shield className="h-4 w-4 text-red-600" />
                    </div>
                    <h3 className="text-lg font-semibold text-slate-900">Security Configuration</h3>
                </div>
                <p className="text-sm text-slate-600 ml-11">
                    Configure rate limiting and connection security for production deployments.
                </p>
            </div>

            <div className="space-y-6">
                {/* Rate Limit Per Minute */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Rate Limit (connections per minute)
                    </label>
                    <input
                        type="number"
                        {...register('security.rate_limit_per_min', { valueAsNumber: true })}
                        placeholder="5"
                        min="1"
                        max="1000"
                        className={getInputClass('security.rate_limit_per_min')}
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Maximum connections allowed per minute from each IP address. Helps prevent abuse and resource exhaustion.
                    </p>
                </div>

                {/* Security Notice */}
                <div className="bg-amber-50 border border-amber-200 rounded-lg p-4">
                    <div className="flex items-start space-x-2">
                        <Shield className="h-5 w-5 text-amber-600 mt-0.5" />
                        <div>
                            <h4 className="text-sm font-medium text-amber-800">Security Configuration Simplified</h4>
                            <p className="text-sm text-amber-700 mt-1">
                                Each configuration file now connects to one specific peer IP address.
                                This simplified approach enhances security by requiring explicit peer-to-peer configurations.
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
} 