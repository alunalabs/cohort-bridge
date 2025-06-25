'use client';

import { useFormContext } from 'react-hook-form';
import { Clock } from 'lucide-react';

export default function TimeoutsSection() {
    const { register } = useFormContext();

    return (
        <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-6">
            <div className="mb-6">
                <div className="flex items-center space-x-3 mb-2">
                    <div className="w-8 h-8 bg-orange-100 rounded-lg flex items-center justify-center">
                        <Clock className="h-4 w-4 text-orange-600" />
                    </div>
                    <h3 className="text-lg font-semibold text-slate-900">Timeout Configuration</h3>
                </div>
                <p className="text-sm text-slate-600 ml-11">
                    Configure network timeouts to handle connection issues and ensure reliable communication.
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
                        {...register('timeouts.connection', { valueAsNumber: true })}
                        placeholder="10"
                        min="1"
                        max="300"
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Time to wait when establishing initial connection to peer
                    </p>
                </div>

                {/* Read Timeout */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Read Timeout (seconds)
                    </label>
                    <input
                        type="number"
                        {...register('timeouts.read', { valueAsNumber: true })}
                        placeholder="30"
                        min="1"
                        max="600"
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Time to wait for data from established connection
                    </p>
                </div>

                {/* Write Timeout */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Write Timeout (seconds)
                    </label>
                    <input
                        type="number"
                        {...register('timeouts.write', { valueAsNumber: true })}
                        placeholder="30"
                        min="1"
                        max="600"
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Time to wait when sending data to peer
                    </p>
                </div>

                {/* Handshake Timeout */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Handshake Timeout (seconds)
                    </label>
                    <input
                        type="number"
                        {...register('timeouts.handshake', { valueAsNumber: true })}
                        placeholder="15"
                        min="1"
                        max="120"
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Time to wait for TLS handshake completion
                    </p>
                </div>

                {/* Keep Alive */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Keep Alive Interval (seconds)
                    </label>
                    <input
                        type="number"
                        {...register('timeouts.keep_alive', { valueAsNumber: true })}
                        placeholder="60"
                        min="10"
                        max="300"
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Interval for sending keep-alive packets to maintain connection
                    </p>
                </div>
            </div>
        </div>
    );
} 