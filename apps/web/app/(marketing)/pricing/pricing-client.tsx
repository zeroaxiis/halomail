"use client";

import { useEffect, useState } from "react";
import { Sparkles, Server } from "lucide-react";
import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";

const PRICING = {
  Free: { monthly: { INR: 0, USD: 0, EUR: 0, GBP: 0, CAD: 0, AUD: 0, SGD: 0, AED: 0 }, yearly: { INR: 0, USD: 0, EUR: 0, GBP: 0, CAD: 0, AUD: 0, SGD: 0, AED: 0 } },
  Pro: {
    monthly: { INR: 299, USD: 3.5, EUR: 3.0, GBP: 2.5, CAD: 5.0, AUD: 5.0, SGD: 4.5, AED: 12.0 },
    yearly: { INR: 2999, USD: 33.0, EUR: 29.5, GBP: 25.0, CAD: 46.5, AUD: 47.0, SGD: 42.0, AED: 119.5 }
  },
  Business: {
    monthly: { INR: 799, USD: 9.0, EUR: 8.0, GBP: 7.0, CAD: 12.5, AUD: 12.5, SGD: 11.5, AED: 32.0 },
    yearly: { INR: 7999, USD: 87.0, EUR: 78.0, GBP: 66.0, CAD: 124.0, AUD: 125.0, SGD: 111.5, AED: 319.0 }
  }
};

const euroCountries = ["AT", "BE", "CY", "EE", "FI", "FR", "DE", "GR", "IE", "IT", "LV", "LT", "LU", "MT", "NL", "PT", "SK", "SI", "ES", "HR"];

function getCurrencyForCountry(countryCode: string) {
  if (countryCode === "IN") return "INR";
  if (countryCode === "US") return "USD";
  if (countryCode === "GB") return "GBP";
  if (countryCode === "CA") return "CAD";
  if (countryCode === "AU") return "AUD";
  if (countryCode === "SG") return "SGD";
  if (countryCode === "AE") return "AED";
  if (euroCountries.includes(countryCode)) return "EUR";
  return "USD";
}

const currencies = ["INR", "USD", "EUR", "GBP", "CAD", "AUD", "SGD", "AED"];

function formatPrice(currency: string, amount: number) {
  if (amount === 0) return `${currency === "INR" ? "₹" : currency === "EUR" ? "€" : currency === "GBP" ? "£" : currency + " "}0`;
  
  let formatted = amount.toString();
  if (currency !== "INR") {
    formatted = amount.toFixed(1);
  }

  const prefix = currency === "INR" ? "₹" : currency === "EUR" ? "€" : currency === "GBP" ? "£" : `${currency} `;
  return `${prefix}${formatted}`;
}

export function PricingClient({ initialCountryCode = "US" }: { initialCountryCode?: string }) {
  const [billingCycle, setBillingCycle] = useState<"monthly" | "yearly">("monthly");
  const [currency, setCurrency] = useState(getCurrencyForCountry(initialCountryCode));
  const [isFetched, setIsFetched] = useState(false);
  const [isProcessing, setIsProcessing] = useState<string | null>(null);

  useEffect(() => {
    // If Vercel headers are missing (e.g. running locally), detect via API
    if (!isFetched) {
      async function detectLocation() {
        try {
          const res = await fetch("https://get.geojs.io/v1/ip/country.json");
          const data = await res.json();
          setCurrency(getCurrencyForCountry(data.country));
        } catch (err) {
          console.error("Failed to detect location", err);
        } finally {
          setIsFetched(true);
        }
      }
      detectLocation();
    }
  }, [isFetched]);

  const loadRazorpayScript = () => {
    return new Promise((resolve) => {
      if ((window as any).Razorpay) return resolve(true);
      const script = document.createElement("script");
      script.src = "https://checkout.razorpay.com/v1/checkout.js";
      script.onload = () => resolve(true);
      script.onerror = () => resolve(false);
      document.body.appendChild(script);
    });
  };

  const handleCheckout = async (planName: "Pro" | "Business") => {
    setIsProcessing(planName);
    const amount = PRICING[planName][billingCycle][currency as keyof typeof PRICING.Pro.monthly];

    const isLoaded = await loadRazorpayScript();
    if (!isLoaded) {
      alert("Failed to load Razorpay SDK. Please check your connection.");
      setIsProcessing(null);
      return;
    }

    try {
      const res = await fetch("/api/razorpay", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ amount, currency, plan: planName, billingCycle })
      });
      const order = await res.json();

      if (order.error) {
        alert("Failed to create order: " + order.error);
        setIsProcessing(null);
        return;
      }

      const options = {
        key: process.env.NEXT_PUBLIC_RAZERPAY_LIVE_API_KEY || "rzp_live_Tl8rjPPjJnFYHq",
        amount: order.amount,
        currency: order.currency,
        name: "ZeroAxiis",
        description: `Upgrade to ${planName} Plan (${billingCycle})`,
        order_id: order.id,
        handler: function (response: any) {
          alert(`Payment successful! Payment ID: ${response.razorpay_payment_id}`);
        },
        theme: { color: "#000000" } // Dark theme
      };

      const rzp = new (window as any).Razorpay(options);
      rzp.on("payment.failed", function (response: any) {
        alert("Payment failed: " + response.error.description);
      });
      rzp.open();
    } catch (err) {
      console.error(err);
      alert("Something went wrong during checkout.");
    } finally {
      setIsProcessing(null);
    }
  };

  const getSavings = (planName: "Pro" | "Business") => {
    const m = PRICING[planName].monthly[currency as keyof typeof PRICING.Pro.monthly];
    const y = PRICING[planName].yearly[currency as keyof typeof PRICING.Pro.yearly];
    if (!m || !y) return 0;
    const savings = 1 - (y / (12 * m));
    return Math.floor(savings * 100);
  };

  return (
    <div className="mt-14">
      <div className="flex justify-center items-center mb-12">
        <div className="flex items-center gap-1 rounded-full border border-border p-1 bg-card/60 backdrop-blur">
          <button
            onClick={() => setBillingCycle("monthly")}
            className={`rounded-full px-6 py-2.5 text-sm font-medium transition-colors ${billingCycle === "monthly" ? "bg-primary text-primary-foreground shadow-sm" : "text-muted-foreground hover:text-foreground"}`}
          >
            Monthly
          </button>
          <button
            onClick={() => setBillingCycle("yearly")}
            className={`rounded-full px-6 py-2.5 text-sm font-medium transition-colors ${billingCycle === "yearly" ? "bg-primary text-primary-foreground shadow-sm" : "text-muted-foreground hover:text-foreground"}`}
          >
            Yearly
          </button>
        </div>
      </div>

      <div className="mx-auto mt-14 grid max-w-6xl gap-5 md:grid-cols-2 xl:grid-cols-4">
        {/* Free Plan */}
        <Card className="relative p-7 flex flex-col bg-card/40 backdrop-blur">
          <div className="flex items-center gap-2">
            <Server className="size-4 text-muted-foreground" />
            <span className="font-medium text-lg">Free</span>
          </div>
          <div className="mt-4 flex flex-col min-h-[64px]">
            <div className="flex items-baseline gap-1.5">
              <span className="text-4xl font-semibold tracking-tight">{formatPrice(currency, 0)}</span>
            </div>
          </div>
          <p className="mt-3 min-h-[48px] text-sm text-muted-foreground">Personal sites. 1 website, 3 forms, 100 submissions and 10 meetings / month.</p>
          <Button asChild className="mt-6 w-full" size="lg" variant="outline">
            <Link href="/register">Start Free</Link>
          </Button>
        </Card>

        {/* Pro Plan */}
        <Card className="relative p-7 flex flex-col border-brand/40 shadow-xl bg-card">
          <span className="absolute -top-3 left-1/2 -translate-x-1/2">
            <Badge variant="brand" className="whitespace-nowrap">Recommended</Badge>
          </span>
          <div className="flex items-center gap-2">
            <Sparkles className="size-4 text-brand" />
            <span className="font-medium text-lg">Pro</span>
          </div>
          <div className="mt-4 flex flex-col min-h-[64px] justify-center">
            <div className="flex items-baseline gap-1.5">
              <span className="text-4xl font-semibold tracking-tight">
                {formatPrice(currency, PRICING.Pro[billingCycle][currency as keyof typeof PRICING.Pro.monthly])}
              </span>
              <span className="text-sm text-muted-foreground">
                {billingCycle === "monthly" ? "/ month" : "/ year"}
              </span>
            </div>
            {billingCycle === "yearly" && (
              <span className="text-xs text-brand font-medium mt-1">Save {getSavings("Pro")}% with yearly</span>
            )}
          </div>
          <p className="mt-3 min-h-[48px] text-sm text-muted-foreground">Freelancers and professionals. 1,000 submissions, 100 meetings, automation.</p>
          <Button 
            className="mt-6 w-full" 
            size="lg" 
            variant="brand"
            onClick={() => handleCheckout("Pro")}
            disabled={isProcessing === "Pro"}
          >
            {isProcessing === "Pro" ? "Processing..." : "Get Pro"}
          </Button>
        </Card>

        {/* Business Plan */}
        <Card className="relative p-7 flex flex-col bg-card/40 backdrop-blur">
          <div className="flex items-center gap-2">
            <Server className="size-4 text-muted-foreground" />
            <span className="font-medium text-lg">Business</span>
          </div>
          <div className="mt-4 flex flex-col min-h-[64px] justify-center">
            <div className="flex items-baseline gap-1.5">
              <span className="text-4xl font-semibold tracking-tight">
                {formatPrice(currency, PRICING.Business[billingCycle][currency as keyof typeof PRICING.Business.monthly])}
              </span>
              <span className="text-sm text-muted-foreground">
                {billingCycle === "monthly" ? "/ month" : "/ year"}
              </span>
            </div>
            {billingCycle === "yearly" && (
              <span className="text-xs text-brand font-medium mt-1">Save {getSavings("Business")}% with yearly</span>
            )}
          </div>
          <p className="mt-3 min-h-[48px] text-sm text-muted-foreground">Growing teams and agencies. 3 users, 2,500 submissions, 250 meetings.</p>
          <Button 
            className="mt-6 w-full" 
            size="lg" 
            variant="outline"
            onClick={() => handleCheckout("Business")}
            disabled={isProcessing === "Business"}
          >
            {isProcessing === "Business" ? "Processing..." : "Get Business"}
          </Button>
        </Card>

        {/* Enterprise Plan */}
        <Card className="relative p-7 flex flex-col bg-card/40 backdrop-blur">
          <div className="flex items-center gap-2">
            <Server className="size-4 text-muted-foreground" />
            <span className="font-medium text-lg">Enterprise</span>
          </div>
          <div className="mt-4 flex flex-col min-h-[64px] justify-center">
            <div className="flex items-baseline gap-1.5">
              <span className="text-4xl font-semibold tracking-tight">Custom</span>
            </div>
          </div>
          <p className="mt-3 min-h-[48px] text-sm text-muted-foreground">Larger organizations. Negotiated usage, security, retention and support.</p>
          <Button asChild className="mt-6 w-full" size="lg" variant="outline">
            <Link href="/contact-sales">Contact Sales</Link>
          </Button>
        </Card>
      </div>

      <div className="mt-10 text-center text-sm text-muted-foreground">
        Prices shown in{" "}
        <select 
          value={currency} 
          onChange={(e) => setCurrency(e.target.value)}
          className="bg-transparent font-medium text-foreground underline decoration-muted-foreground/50 underline-offset-4 hover:decoration-foreground cursor-pointer focus:outline-none"
        >
          {currencies.map(c => <option key={c} value={c}>{c}</option>)}
        </select>
        {" "}based on your location.
      </div>
    </div>
  );
}
