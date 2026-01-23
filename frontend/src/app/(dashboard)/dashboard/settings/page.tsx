'use client';

import { useState } from 'react';
import { Header } from '@/components/layout/header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Switch } from '@/components/ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Store,
  Receipt,
  Bell,
  Palette,
  Save,
  Loader2,
} from 'lucide-react';

export default function SettingsPage() {
  const [isSaving, setIsSaving] = useState(false);

  // Store settings
  const [storeSettings, setStoreSettings] = useState({
    storeName: 'DashPoint Store',
    address: '',
    phone: '',
    taxRate: 0,
    currency: 'IDR',
  });

  // Receipt settings
  const [receiptSettings, setReceiptSettings] = useState({
    showLogo: true,
    headerText: 'Thank you for shopping with us!',
    footerText: 'Please come again!',
    showTaxDetails: true,
  });

  // Notification settings
  const [notificationSettings, setNotificationSettings] = useState({
    lowStockAlerts: true,
    lowStockThreshold: 10,
    dailySummaryEmail: false,
  });

  // Appearance settings
  const [appearanceSettings, setAppearanceSettings] = useState({
    theme: 'system',
    compactMode: false,
  });

  const handleSave = async () => {
    setIsSaving(true);
    // Simulate save - in production this would call the backend
    await new Promise((resolve) => setTimeout(resolve, 1000));
    setIsSaving(false);
  };

  return (
    <div className="flex flex-col h-full">
      <Header title="Settings" />

      <div className="flex-1 p-6 overflow-auto">
        <div className="max-w-3xl mx-auto space-y-6">
          {/* Store Information */}
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Store className="h-5 w-5 text-primary" />
                <div>
                  <CardTitle>Store Information</CardTitle>
                  <CardDescription>Basic information about your store</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-2">
                <Label htmlFor="storeName">Store Name</Label>
                <Input
                  id="storeName"
                  value={storeSettings.storeName}
                  onChange={(e) =>
                    setStoreSettings({ ...storeSettings, storeName: e.target.value })
                  }
                  placeholder="Enter store name"
                />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="address">Address</Label>
                <Input
                  id="address"
                  value={storeSettings.address}
                  onChange={(e) =>
                    setStoreSettings({ ...storeSettings, address: e.target.value })
                  }
                  placeholder="Store address"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="grid gap-2">
                  <Label htmlFor="phone">Phone Number</Label>
                  <Input
                    id="phone"
                    value={storeSettings.phone}
                    onChange={(e) =>
                      setStoreSettings({ ...storeSettings, phone: e.target.value })
                    }
                    placeholder="+62 xxx xxxx xxxx"
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="taxRate">Tax Rate (%)</Label>
                  <Input
                    id="taxRate"
                    type="number"
                    value={storeSettings.taxRate}
                    onChange={(e) =>
                      setStoreSettings({
                        ...storeSettings,
                        taxRate: parseFloat(e.target.value) || 0,
                      })
                    }
                    placeholder="0"
                  />
                </div>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="currency">Currency</Label>
                <Select
                  value={storeSettings.currency}
                  onValueChange={(value) =>
                    setStoreSettings({ ...storeSettings, currency: value })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="IDR">Indonesian Rupiah (IDR)</SelectItem>
                    <SelectItem value="USD">US Dollar (USD)</SelectItem>
                    <SelectItem value="SGD">Singapore Dollar (SGD)</SelectItem>
                    <SelectItem value="MYR">Malaysian Ringgit (MYR)</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          </Card>

          {/* Receipt Settings */}
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Receipt className="h-5 w-5 text-primary" />
                <div>
                  <CardTitle>Receipt Settings</CardTitle>
                  <CardDescription>Customize how receipts are printed</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <Label>Show Logo</Label>
                  <p className="text-sm text-muted-foreground">
                    Display store logo on receipts
                  </p>
                </div>
                <Switch
                  checked={receiptSettings.showLogo}
                  onCheckedChange={(checked) =>
                    setReceiptSettings({ ...receiptSettings, showLogo: checked })
                  }
                />
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <Label>Show Tax Details</Label>
                  <p className="text-sm text-muted-foreground">
                    Display tax breakdown on receipts
                  </p>
                </div>
                <Switch
                  checked={receiptSettings.showTaxDetails}
                  onCheckedChange={(checked) =>
                    setReceiptSettings({ ...receiptSettings, showTaxDetails: checked })
                  }
                />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="headerText">Header Message</Label>
                <Input
                  id="headerText"
                  value={receiptSettings.headerText}
                  onChange={(e) =>
                    setReceiptSettings({ ...receiptSettings, headerText: e.target.value })
                  }
                  placeholder="Thank you for shopping with us!"
                />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="footerText">Footer Message</Label>
                <Input
                  id="footerText"
                  value={receiptSettings.footerText}
                  onChange={(e) =>
                    setReceiptSettings({ ...receiptSettings, footerText: e.target.value })
                  }
                  placeholder="Please come again!"
                />
              </div>
            </CardContent>
          </Card>

          {/* Notification Settings */}
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Bell className="h-5 w-5 text-primary" />
                <div>
                  <CardTitle>Notifications</CardTitle>
                  <CardDescription>Configure alerts and notifications</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <Label>Low Stock Alerts</Label>
                  <p className="text-sm text-muted-foreground">
                    Show alerts when products are running low
                  </p>
                </div>
                <Switch
                  checked={notificationSettings.lowStockAlerts}
                  onCheckedChange={(checked) =>
                    setNotificationSettings({
                      ...notificationSettings,
                      lowStockAlerts: checked,
                    })
                  }
                />
              </div>

              {notificationSettings.lowStockAlerts && (
                <div className="grid gap-2">
                  <Label htmlFor="lowStockThreshold">Low Stock Threshold</Label>
                  <Input
                    id="lowStockThreshold"
                    type="number"
                    value={notificationSettings.lowStockThreshold}
                    onChange={(e) =>
                      setNotificationSettings({
                        ...notificationSettings,
                        lowStockThreshold: parseInt(e.target.value) || 0,
                      })
                    }
                    placeholder="10"
                  />
                  <p className="text-xs text-muted-foreground">
                    Alert when product quantity falls below this number
                  </p>
                </div>
              )}

              <div className="flex items-center justify-between">
                <div>
                  <Label>Daily Summary Email</Label>
                  <p className="text-sm text-muted-foreground">
                    Receive daily sales summary via email
                  </p>
                </div>
                <Switch
                  checked={notificationSettings.dailySummaryEmail}
                  onCheckedChange={(checked) =>
                    setNotificationSettings({
                      ...notificationSettings,
                      dailySummaryEmail: checked,
                    })
                  }
                />
              </div>
            </CardContent>
          </Card>

          {/* Appearance Settings */}
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Palette className="h-5 w-5 text-primary" />
                <div>
                  <CardTitle>Appearance</CardTitle>
                  <CardDescription>Customize the look and feel</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-2">
                <Label htmlFor="theme">Theme</Label>
                <Select
                  value={appearanceSettings.theme}
                  onValueChange={(value) =>
                    setAppearanceSettings({ ...appearanceSettings, theme: value })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="light">Light</SelectItem>
                    <SelectItem value="dark">Dark</SelectItem>
                    <SelectItem value="system">System Default</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <Label>Compact Mode</Label>
                  <p className="text-sm text-muted-foreground">
                    Use smaller spacing and font sizes
                  </p>
                </div>
                <Switch
                  checked={appearanceSettings.compactMode}
                  onCheckedChange={(checked) =>
                    setAppearanceSettings({ ...appearanceSettings, compactMode: checked })
                  }
                />
              </div>
            </CardContent>
          </Card>

          {/* Save Button */}
          <div className="flex justify-end">
            <Button onClick={handleSave} disabled={isSaving} size="lg">
              {isSaving ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Saving...
                </>
              ) : (
                <>
                  <Save className="h-4 w-4 mr-2" />
                  Save Settings
                </>
              )}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
