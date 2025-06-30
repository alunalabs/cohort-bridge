'use client';

import { useFormContext } from 'react-hook-form';
import { Clock } from 'lucide-react';

interface TimeoutsSectionProps {
    missingFields?: string[];
}

export default function TimeoutsSection({ missingFields = [] }: TimeoutsSectionProps) {
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
                    <div className="w-8 h-8 bg-yellow-100 rounded-lg flex items-center justify-center">
                        <Clock className="h-4 w-4 text-yellow-600" />
                    </div>
                    <h3 className="text-lg font-semibold text-slate-900">Timeouts Configuration</h3>
                </div>
                <p className="text-sm text-slate-600 ml-11">
                    Configure connection and operation timeouts to prevent hanging connections and ensure reliable operations.
                </p>
            </div>

            <div className="space-y-6">
                {/* Read Timeout */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Read Timeout (seconds)
                    </label>
                    <input
                        type="number"
                        {...register('timeouts.read_timeout', { valueAsNumber: true })}
                        placeholder="30"
                        min="1"
                        className={getInputClass('timeouts.read_timeout')}
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Maximum time to wait for data from peer. 30-60 seconds recommended for large datasets.
                    </p>
                </div>

                {/* Write Timeout */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Write Timeout (seconds)
                    </label>
                    <input
                        type="number"
                        {...register('timeouts.write_timeout', { valueAsNumber: true })}
                        placeholder="30"
                        min="1"
                        className={getInputClass('timeouts.write_timeout')}
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Maximum time to wait when sending data to peer. Should match or exceed read timeout.
                    </p>
                </div>

                {/* Connection Timeout */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Connection Timeout (seconds)
                    </label>
                    <input
                        type="number"
                        {...register('timeouts.connection_timeout', { valueAsNumber: true })}
                        placeholder="10"
                        min="1"
                        className={getInputClass('timeouts.connection_timeout')}
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Maximum time to wait when establishing initial connection. 5-15 seconds typical.
                    </p>
                </div>

                {/* Keep Alive Interval */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Keep Alive Interval (seconds)
                    </label>
                    <input
                        type="number"
                        {...register('timeouts.keep_alive_interval', { valueAsNumber: true })}
                        placeholder="60"
                        min="1"
                        className={getInputClass('timeouts.keep_alive_interval')}
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Frequency of keep-alive messages to maintain connection during long operations.
                    </p>
                </div>
            </div>
        </div>
    );
} 