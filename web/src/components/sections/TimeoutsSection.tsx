'use client';

import { useFormContext } from 'react-hook-form';
import { Clock, Info } from 'lucide-react';

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
                {/* Connection Timeout */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Connection Timeout (seconds)
                    </label>
                    <input
                        type="number"
                        {...register('timeouts.connection_timeout', { valueAsNumber: true })}
                        placeholder="30"
                        min="1"
                        className={getInputClass('timeouts.connection_timeout')}
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Maximum time to wait when establishing initial connection. 10-60 seconds recommended.
                    </p>
                </div>

                {/* Read Timeout */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Read Timeout (seconds)
                    </label>
                    <input
                        type="number"
                        {...register('timeouts.read_timeout', { valueAsNumber: true })}
                        placeholder="60"
                        min="1"
                        className={getInputClass('timeouts.read_timeout')}
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Maximum time to wait for data from peer. 60-180 seconds recommended for large datasets.
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
                        placeholder="60"
                        min="1"
                        className={getInputClass('timeouts.write_timeout')}
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Maximum time to wait when sending data to peer. Should match or exceed read timeout.
                    </p>
                </div>

                {/* Idle Timeout */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Idle Timeout (seconds)
                    </label>
                    <input
                        type="number"
                        {...register('timeouts.idle_timeout', { valueAsNumber: true })}
                        placeholder="300"
                        min="1"
                        className={getInputClass('timeouts.idle_timeout')}
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Maximum idle time before closing connection. 300-600 seconds (5-10 minutes) typical.
                    </p>
                </div>

                {/* Handshake Timeout */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Handshake Timeout (seconds)
                    </label>
                    <input
                        type="number"
                        {...register('timeouts.handshake_timeout', { valueAsNumber: true })}
                        placeholder="30"
                        min="1"
                        className={getInputClass('timeouts.handshake_timeout')}
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Maximum time to wait for protocol handshake completion. 30-60 seconds recommended.
                    </p>
                </div>

                {/* Timeout Guidelines */}
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <div className="flex items-start space-x-3">
                        <Info className="h-5 w-5 text-blue-600 mt-0.5" />
                        <div>
                            <h4 className="text-sm font-medium text-blue-900 mb-2">Timeout Guidelines</h4>
                            <div className="space-y-2 text-sm text-blue-800">
                                <p>
                                    <strong>Small datasets (&lt;10K records):</strong> Use default values (30-60 seconds)
                                </p>
                                <p>
                                    <strong>Large datasets (&gt;100K records):</strong> Increase read/write timeouts to 120-180 seconds
                                </p>
                                <p>
                                    <strong>Secure protocols:</strong> May require longer handshake timeouts (45-60 seconds)
                                </p>
                                <p className="text-xs italic">
                                    Network latency and system performance affect optimal timeout values.
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
} 