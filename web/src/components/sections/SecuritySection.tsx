'use client';

import { useFormContext } from 'react-hook-form';
import { Shield, Plus, Minus } from 'lucide-react';

export default function SecuritySection() {
    const { register, watch, setValue } = useFormContext();

    const allowedIps = watch('security.allowed_ips') || [];

    const addIp = () => {
        setValue('security.allowed_ips', [...allowedIps, '']);
    };

    const removeIp = (index: number) => {
        const newIps = allowedIps.filter((_: any, i: number) => i !== index);
        setValue('security.allowed_ips', newIps);
    };

    const updateIp = (index: number, value: string) => {
        const newIps = [...allowedIps];
        newIps[index] = value;
        setValue('security.allowed_ips', newIps);
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
                    Control network access, rate limiting, and connection security for production deployments.
                </p>
            </div>

            <div className="space-y-6">
                {/* Allowed IPs */}
                <div>
                    <div className="flex items-center justify-between mb-3">
                        <label className="block text-sm font-medium text-slate-700">
                            Allowed IP Addresses
                        </label>
                        <button
                            type="button"
                            onClick={addIp}
                            className="flex items-center space-x-1 text-blue-600 hover:text-blue-700 text-sm"
                        >
                            <Plus className="h-4 w-4" />
                            <span>Add IP</span>
                        </button>
                    </div>

                    <div className="space-y-2">
                        {allowedIps.map((ip: string, index: number) => (
                            <div key={index} className="flex items-center space-x-2">
                                <input
                                    type="text"
                                    value={ip}
                                    onChange={(e) => updateIp(index, e.target.value)}
                                    placeholder="192.168.1.0/24 or 10.0.0.1"
                                    className="flex-1 px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                />
                                {allowedIps.length > 1 && (
                                    <button
                                        type="button"
                                        onClick={() => removeIp(index)}
                                        className="p-2 text-red-600 hover:text-red-700"
                                    >
                                        <Minus className="h-4 w-4" />
                                    </button>
                                )}
                            </div>
                        ))}
                    </div>
                    <p className="mt-1 text-xs text-slate-600">
                        Whitelist trusted IPs/networks (e.g., 192.168.1.0/24 for local network, 10.0.0.0/8 for VPN)
                    </p>
                </div>

                {/* Require IP Check */}
                <div>
                    <label className="flex items-center space-x-3">
                        <input
                            type="checkbox"
                            {...register('security.require_ip_check')}
                            className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                        />
                        <span className="text-sm font-medium text-slate-700">Require IP Address Check</span>
                    </label>
                    <p className="mt-1 text-xs text-slate-600 ml-6">
                        Enforce IP whitelist validation for all connections. Disable for development only.
                    </p>
                </div>

                {/* Max Connections */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Maximum Concurrent Connections
                    </label>
                    <input
                        type="number"
                        {...register('security.max_connections', { valueAsNumber: true })}
                        placeholder="5"
                        min="1"
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Prevent resource exhaustion. 5-10 connections typical for two-party linkage.
                    </p>
                </div>

                {/* Rate Limit */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Rate Limit (requests per minute)
                    </label>
                    <input
                        type="number"
                        {...register('security.rate_limit.requests_per_minute', { valueAsNumber: true })}
                        placeholder="100"
                        min="1"
                        max="10000"
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Maximum requests allowed per minute from each IP address
                    </p>
                </div>

                {/* Rate Limit - Burst Size */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Burst Size
                    </label>
                    <input
                        type="number"
                        {...register('security.rate_limit.burst_size', { valueAsNumber: true })}
                        placeholder="10"
                        min="1"
                        max="1000"
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Maximum requests allowed in a burst
                    </p>
                </div>
            </div>
        </div>
    );
} 