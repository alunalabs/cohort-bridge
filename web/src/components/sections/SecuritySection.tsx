'use client';

import { useFormContext } from 'react-hook-form';
import { Shield, Plus, Minus, AlertTriangle } from 'lucide-react';
import { useEffect } from 'react';

interface SecuritySectionProps {
    missingFields?: string[];
}

export default function SecuritySection({ missingFields = [] }: SecuritySectionProps) {
    const { register, watch, setValue } = useFormContext();

    const allowedIps = watch('_ui_allowed_ips') || [];
    const requireIpCheck = watch('security.require_ip_check');

    // Initialize with default IPs if none exist
    useEffect(() => {
        if (allowedIps.length === 0) {
            setValue('_ui_allowed_ips', ['127.0.0.1', '::1']);
            setValue('security.allowed_ips', ['127.0.0.1', '::1']);
        }
    }, [allowedIps.length, setValue]);

    // Update security.allowed_ips when UI changes
    const updateAllowedIps = (newIps: string[]) => {
        const validIps = newIps.filter(ip => ip.trim() !== '');
        setValue('_ui_allowed_ips', newIps);
        setValue('security.allowed_ips', validIps);
    };

    const addIpAddress = () => {
        const newIps = [...allowedIps, ''];
        updateAllowedIps(newIps);
    };

    const removeIpAddress = (index: number) => {
        const newIps = allowedIps.filter((_: string, i: number) => i !== index);
        updateAllowedIps(newIps);
    };

    const updateIpAddress = (index: number, value: string) => {
        const newIps = [...allowedIps];
        newIps[index] = value;
        updateAllowedIps(newIps);
    };

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
                    Configure access controls, rate limiting, and IP restrictions for production security.
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

                {/* Maximum Connections */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Maximum Concurrent Connections
                    </label>
                    <input
                        type="number"
                        {...register('security.max_connections', { valueAsNumber: true })}
                        placeholder="5"
                        min="1"
                        max="100"
                        className={getInputClass('security.max_connections')}
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Maximum number of simultaneous connections allowed. Higher values use more system resources.
                    </p>
                </div>

                {/* IP Access Control */}
                <div>
                    <div className="flex items-center justify-between mb-3">
                        <div>
                            <label className="block text-sm font-medium text-slate-700">
                                IP Access Control
                            </label>
                            <p className="text-xs text-slate-600 mt-1">
                                Control which IP addresses can connect to your system
                            </p>
                        </div>
                        <label className="flex items-center space-x-2">
                            <input
                                type="checkbox"
                                {...register('security.require_ip_check')}
                                className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                            />
                            <span className="text-sm font-medium text-slate-700">Enable IP restrictions</span>
                        </label>
                    </div>

                    {requireIpCheck && (
                        <div className="border border-slate-200 rounded-lg p-4 bg-slate-50">
                            <div className="flex items-center justify-between mb-3">
                                <h5 className="text-sm font-medium text-slate-700">Allowed IP Addresses</h5>
                                <button
                                    type="button"
                                    onClick={addIpAddress}
                                    className="flex items-center space-x-1 text-blue-600 hover:text-blue-700 text-sm font-medium"
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
                                            onChange={(e) => updateIpAddress(index, e.target.value)}
                                            placeholder="192.168.1.100 or 192.168.1.0/24"
                                            className="flex-1 px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 bg-white text-sm"
                                        />
                                        {allowedIps.length > 1 && (
                                            <button
                                                type="button"
                                                onClick={() => removeIpAddress(index)}
                                                className="p-2 text-red-600 hover:text-red-700 hover:bg-red-50 rounded-lg transition-colors"
                                            >
                                                <Minus className="h-4 w-4" />
                                            </button>
                                        )}
                                    </div>
                                ))}
                            </div>

                            <div className="mt-3 text-xs text-slate-600">
                                <p className="mb-1"><strong>Examples:</strong></p>
                                <ul className="space-y-1 ml-4">
                                    <li>• <code>127.0.0.1</code> - localhost IPv4</li>
                                    <li>• <code>::1</code> - localhost IPv6</li>
                                    <li>• <code>192.168.1.100</code> - specific IP address</li>
                                    <li>• <code>192.168.1.0/24</code> - entire subnet</li>
                                </ul>
                            </div>
                        </div>
                    )}

                    {!requireIpCheck && (
                        <div className="bg-amber-50 border border-amber-200 rounded-lg p-3">
                            <div className="flex items-start space-x-2">
                                <AlertTriangle className="h-4 w-4 text-amber-600 mt-0.5" />
                                <div className="text-xs text-amber-700">
                                    <strong>Warning:</strong> IP restrictions are disabled. Any IP address can connect to your system.
                                    Enable IP restrictions for production deployments.
                                </div>
                            </div>
                        </div>
                    )}
                </div>

                {/* Security Notice */}
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <div className="flex items-start space-x-2">
                        <Shield className="h-5 w-5 text-blue-600 mt-0.5" />
                        <div>
                            <h4 className="text-sm font-medium text-blue-800">Security Best Practices</h4>
                            <ul className="text-sm text-blue-700 mt-2 space-y-1">
                                <li>• Enable IP restrictions for production deployments</li>
                                <li>• Use conservative rate limits to prevent abuse</li>
                                <li>• Monitor audit logs for suspicious activity</li>
                                <li>• Coordinate security settings with peer organizations</li>
                            </ul>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
} 