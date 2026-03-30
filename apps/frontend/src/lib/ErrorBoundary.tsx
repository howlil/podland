"use client"

import { Component, ReactNode, ErrorInfo } from "react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { AlertCircle } from "lucide-react"
import { Button } from "@/components/ui/button"

interface Props {
  children: ReactNode
  fallback?: ReactNode
  onError?: (error: Error, errorInfo: ErrorInfo) => void
}

interface State {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error("ErrorBoundary caught an error:", error, errorInfo)
    this.props.onError?.(error, errorInfo)
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null })
    window.location.reload()
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }

      return (
        <div className="min-h-screen flex items-center justify-center p-4 bg-gray-50 dark:bg-gray-900">
          <Alert variant="destructive" className="max-w-md">
            <AlertCircle className="h-5 w-5" />
            <AlertTitle>Something went wrong</AlertTitle>
            <AlertDescription className="mt-2">
              <p className="text-sm mb-4">
                {this.state.error?.message || "An unexpected error occurred"}
              </p>
              <div className="flex gap-2 mt-4">
                <Button onClick={this.handleReset} variant="default" size="sm">
                  Reload Page
                </Button>
                <Button
                  onClick={() => window.history.back()}
                  variant="outline"
                  size="sm"
                >
                  Go Back
                </Button>
              </div>
            </AlertDescription>
          </Alert>
        </div>
      )
    }

    return this.props.children
  }
}

export default ErrorBoundary
